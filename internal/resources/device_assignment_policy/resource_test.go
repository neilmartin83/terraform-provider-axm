// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package device_assignment_policy

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	providerschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

type assignmentCall struct{ id, destination string }

type fakePolicyAPI struct {
	mu             sync.Mutex
	snapshot       Snapshot
	calls          []assignmentCall
	snapshotCount  int
	beforeSnapshot func(*fakePolicyAPI) error
	readDevice     func(Device) (Device, error)
	readServer     func(Server) (Server, error)
	write          func(*fakePolicyAPI, string, string) error
}

func (f *fakePolicyAPI) Snapshot(context.Context) (Snapshot, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.snapshotCount++
	if f.beforeSnapshot != nil {
		if err := f.beforeSnapshot(f); err != nil {
			return Snapshot{}, err
		}
	}
	return Snapshot{Servers: append([]Server(nil), f.snapshot.Servers...), Devices: append([]Device(nil), f.snapshot.Devices...)}, nil
}

func (f *fakePolicyAPI) Device(_ context.Context, id string) (Device, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, device := range f.snapshot.Devices {
		if device.ID == id {
			if f.readDevice != nil {
				return f.readDevice(device)
			}
			return device, nil
		}
	}
	return Device{}, errors.New("private device lookup error")
}

func (f *fakePolicyAPI) Server(_ context.Context, id string) (Server, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, server := range f.snapshot.Servers {
		if server.ID == id {
			if f.readServer != nil {
				return f.readServer(server)
			}
			return server, nil
		}
	}
	return Server{}, errors.New("private server lookup error")
}

func (f *fakePolicyAPI) Assign(_ context.Context, id, destination string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, assignmentCall{id, destination})
	if f.write != nil {
		return f.write(f, id, destination)
	}
	f.move(id, destination)
	return nil
}

func (f *fakePolicyAPI) move(id, destination string) {
	for i := range f.snapshot.Devices {
		if f.snapshot.Devices[i].ID == id {
			f.snapshot.Devices[i].ServerID = destination
			f.snapshot.Devices[i].Status = "ASSIGNED"
			if destination == "" {
				f.snapshot.Devices[i].Status = Unassigned
			}
		}
	}
}

func (f *fakePolicyAPI) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.calls)
}

func applyFixture(t *testing.T) (*AssignmentPolicyResource, *fakePolicyAPI, Policy, Plan) {
	t.Helper()
	p, s := testPolicy(), testSnapshot()
	p.AdoptionOnly = false
	p.Assignments = map[string]string{"a": "dev", "b": Unassigned}
	s.Devices = []Device{testDevice("a", "prod"), testDevice("b", "live"), testDevice("extra", "dev"), testDevice("untouched", "prod")}
	api := &fakePolicyAPI{snapshot: s}
	approved, err := BuildPlan(p, s)
	if err != nil {
		t.Fatal(err)
	}
	return &AssignmentPolicyResource{api: api}, api, p, approved
}

func TestApplyUsesSingleDirectDestinationsAndVerifiesConvergence(t *testing.T) {
	r, api, p, approved := applyFixture(t)
	if err := r.apply(context.Background(), p, approved); err != nil {
		t.Fatal(err)
	}
	want := []assignmentCall{{"id-a", "dev"}, {"id-b", ""}, {"id-extra", "prod"}}
	if !reflect.DeepEqual(api.calls, want) {
		t.Fatalf("assignment calls = %#v, want %#v", api.calls, want)
	}
	if api.snapshotCount != 2 {
		t.Fatalf("expected before/after snapshots, got %d", api.snapshotCount)
	}
	final, err := r.prepare(context.Background(), p)
	if err != nil || len(final.Changes) != 0 || !reflect.DeepEqual(final.Observed, p.Assignments) {
		t.Fatalf("post-apply assignments did not converge: %#v, %v", final, err)
	}
}

func TestApplyRejectsStalePlanWithoutWrites(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*fakePolicyAPI)
	}{
		{"new_extra", func(f *fakePolicyAPI) { f.snapshot.Devices = append(f.snapshot.Devices, testDevice("new", "live")) }},
		{"device_moved", func(f *fakePolicyAPI) { f.move("id-a", "live") }},
		{"device_disappeared", func(f *fakePolicyAPI) { f.snapshot.Devices = f.snapshot.Devices[1:] }},
		{"device_identity_changed", func(f *fakePolicyAPI) { f.snapshot.Devices[0].ID = "changed-id" }},
		{"same_count_extra_swap", func(f *fakePolicyAPI) { f.snapshot.Devices[2] = testDevice("replacement", "dev") }},
		{"server_renamed", func(f *fakePolicyAPI) { f.snapshot.Servers[1].Name = "private changed name" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r, api, p, approved := applyFixture(t)
			test.mutate(api)
			err := r.apply(context.Background(), p, approved)
			if err == nil || api.callCount() != 0 {
				t.Fatalf("stale plan must fail before writes: %v, calls=%d", err, api.callCount())
			}
			assertPrivateError(t, err)
		})
	}
}

func TestApplyIgnoresUnrelatedDefaultArrivals(t *testing.T) {
	r, api, p, approved := applyFixture(t)
	api.snapshot.Devices = append(api.snapshot.Devices, testDevice("new-default", "prod"))
	if err := r.apply(context.Background(), p, approved); err != nil {
		t.Fatalf("unmanaged default arrival blocked unchanged plan: %v", err)
	}
	for _, call := range api.calls {
		if call.id == "id-new-default" {
			t.Fatal("wrote an unrelated default device")
		}
	}
}

func TestApplyEnforcesAdoptionAndLimitsBeforeWrites(t *testing.T) {
	for _, test := range []struct {
		name     string
		adoption bool
		limit    int64
	}{{"adoption_only", true, 100}, {"zero_limit", false, 0}, {"over_limit", false, 2}} {
		t.Run(test.name, func(t *testing.T) {
			r, api, p, approved := applyFixture(t)
			p.AdoptionOnly, p.MaxChanges = test.adoption, test.limit
			if err := r.apply(context.Background(), p, approved); err == nil || api.callCount() != 0 {
				t.Fatalf("write boundary did not enforce gate: %v", err)
			}
		})
	}
}

func TestApplyRejectsLastMomentDeviceChanges(t *testing.T) {
	tests := []struct {
		name string
		read func(Device) (Device, error)
	}{
		{"read_error", func(d Device) (Device, error) { return d, errors.New("private device error") }},
		{"id_changed", func(d Device) (Device, error) { d.ID = "private-id"; return d, nil }},
		{"serial_changed", func(d Device) (Device, error) { d.SerialNumber = "private-serial"; return d, nil }},
		{"server_changed", func(d Device) (Device, error) { d.ServerID = "live"; return d, nil }},
		{"released", func(d Device) (Device, error) { d.Status = "RELEASED"; return d, nil }},
		{"unknown_status", func(d Device) (Device, error) { d.Status = "private-status"; return d, nil }},
		{"inconsistent_unassigned", func(d Device) (Device, error) { d.Status = Unassigned; return d, nil }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r, api, p, approved := applyFixture(t)
			api.readDevice = test.read
			err := r.apply(context.Background(), p, approved)
			if err == nil || api.callCount() != 0 {
				t.Fatalf("changed device must fail before writes: %v, calls=%d", err, api.callCount())
			}
			assertPrivateError(t, err)
		})
	}
}

func TestApplyRejectsLastMomentTargetChanges(t *testing.T) {
	for _, test := range []struct {
		name string
		read func(Server) (Server, error)
	}{
		{"read_error", func(s Server) (Server, error) { return s, errors.New("private server error") }},
		{"id_changed", func(s Server) (Server, error) { s.ID = "private-id"; return s, nil }},
		{"name_changed", func(s Server) (Server, error) { s.Name = "private-name"; return s, nil }},
		{"type_changed", func(s Server) (Server, error) { s.Type = "APPLE_CONFIGURATOR"; return s, nil }},
	} {
		t.Run(test.name, func(t *testing.T) {
			r, api, p, approved := applyFixture(t)
			api.readServer = test.read
			err := r.apply(context.Background(), p, approved)
			if err == nil || api.callCount() != 0 {
				t.Fatalf("changed target must fail before writes: %v", err)
			}
			assertPrivateError(t, err)
		})
	}
}

func TestApplyStopsOnPartialFailureWithoutRetryOrRollback(t *testing.T) {
	r, api, p, approved := applyFixture(t)
	api.write = func(f *fakePolicyAPI, id, destination string) error {
		if len(f.calls) == 2 {
			return errors.New("private activity timeout")
		}
		f.move(id, destination)
		return nil
	}
	err := r.apply(context.Background(), p, approved)
	if err == nil || api.callCount() != 2 {
		t.Fatalf("must stop after failed second call: %v, calls=%#v", err, api.calls)
	}
	assertPrivateError(t, err)
	if api.snapshot.Devices[0].ServerID != "dev" || api.snapshot.Devices[2].ServerID != "dev" {
		t.Fatal("rolled back the success or applied a later change")
	}
}

func TestApplyFailsPostVerificationWithoutAdditionalWrites(t *testing.T) {
	for _, test := range []struct {
		name  string
		setup func(*fakePolicyAPI)
	}{
		{"write_did_not_converge", func(f *fakePolicyAPI) { f.write = func(*fakePolicyAPI, string, string) error { return nil } }},
		{"new_extra_during_apply", func(f *fakePolicyAPI) {
			f.write = func(f *fakePolicyAPI, id, destination string) error {
				f.move(id, destination)
				if len(f.calls) == 1 {
					f.snapshot.Devices = append(f.snapshot.Devices, testDevice("late-extra", "dev"))
				}
				return nil
			}
		}},
		{"final_snapshot_failed", func(f *fakePolicyAPI) {
			f.beforeSnapshot = func(f *fakePolicyAPI) error {
				if f.snapshotCount == 2 {
					return errors.New("private snapshot failure")
				}
				return nil
			}
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			r, api, p, approved := applyFixture(t)
			test.setup(api)
			err := r.apply(context.Background(), p, approved)
			if err == nil || api.callCount() != len(approved.Changes) {
				t.Fatalf("must fail without unplanned corrective writes: %v, calls=%#v", err, api.calls)
			}
			assertPrivateError(t, err)
		})
	}
}

type policyTestProvider struct{ api *fakePolicyAPI }

func (p policyTestProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName, resp.Version = "axm", "test"
}
func (p policyTestProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = providerschema.Schema{}
}
func (p policyTestProvider) Configure(context.Context, provider.ConfigureRequest, *provider.ConfigureResponse) {
}
func (p policyTestProvider) DataSources(context.Context) []func() datasource.DataSource {
	return nil
}
func (p policyTestProvider) Resources(context.Context) []func() frameworkresource.Resource {
	return []func() frameworkresource.Resource{func() frameworkresource.Resource { return &AssignmentPolicyResource{api: p.api} }}
}

func policyFactories(api *fakePolicyAPI) map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"axm": providerserver.NewProtocol6WithError(policyTestProvider{api: api}),
	}
}

func policyConfig(adoption bool, assignments string) string {
	return fmt.Sprintf(`
provider "axm" {}
resource "axm_device_assignment_policy" "test" {
  assignments = %s
  authoritative_server_ids = ["dev", "live"]
  fallback_server_id = "prod"
  server_names = { prod = "Default MDM", dev = "Development MDM", live = "Live MDM", other = "Other MDM" }
  adoption_only = %t
  max_changes = 3
}
`, assignments, adoption)
}

func checkCalls(api *fakePolicyAPI, want int) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if got := api.callCount(); got != want {
			return fmt.Errorf("assignment write count = %d, want %d", got, want)
		}
		return nil
	}
}

func TestResourceTerraformLifecycle(t *testing.T) {
	s := testSnapshot()
	s.Devices = []Device{testDevice("explicit", "dev"), testDevice("exception", Unassigned), testDevice("unmanaged", "prod")}
	api := &fakePolicyAPI{snapshot: s}
	const address = "axm_device_assignment_policy.test"
	const assignments = `{ explicit = "dev", exception = "UNASSIGNED" }`
	resource.Test(t, resource.TestCase{
		IsUnitTest: true, ProtoV6ProviderFactories: policyFactories(api),
		CheckDestroy: checkCalls(api, 3),
		Steps: []resource.TestStep{
			{
				Config:           policyConfig(true, assignments),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(address, plancheck.ResourceActionCreate)}},
				Check:            resource.ComposeTestCheckFunc(checkCalls(api, 0), resource.TestCheckResourceAttr(address, "id", policyID), resource.TestCheckResourceAttr(address, "device_count", "2")),
			},
			{
				Config: policyConfig(false, assignments),
				Check:  checkCalls(api, 0),
			},
			{
				PreConfig:        func() { api.mu.Lock(); defer api.mu.Unlock(); api.move("id-explicit", "prod") },
				Config:           policyConfig(false, assignments),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(address, plancheck.ResourceActionUpdate)}},
				Check:            resource.ComposeTestCheckFunc(checkCalls(api, 1), resource.TestCheckResourceAttr(address, "observed_assignments.explicit", "dev")),
			},
			{
				PreConfig: func() {
					api.mu.Lock()
					defer api.mu.Unlock()
					api.snapshot.Devices = append(api.snapshot.Devices, testDevice("extra", "live"))
				},
				Config:           policyConfig(false, assignments),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(address, plancheck.ResourceActionUpdate)}},
				Check:            resource.ComposeTestCheckFunc(checkCalls(api, 2), resource.TestCheckNoResourceAttr(address, "observed_assignments.extra")),
			},
			{
				Config: policyConfig(false, `{ explicit = "prod", exception = "UNASSIGNED" }`),
				Check:  resource.ComposeTestCheckFunc(checkCalls(api, 3), resource.TestCheckResourceAttr(address, "observed_assignments.explicit", "prod")),
			},
			{
				Config:           policyConfig(false, `{ explicit = "prod", exception = "UNASSIGNED" }`),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}},
				Check:            checkCalls(api, 3),
			},
		},
	})
}

func TestResourceTerraformRejectsDriftBetweenReadAndModifyPlan(t *testing.T) {
	s := testSnapshot()
	s.Devices = []Device{testDevice("explicit", "dev")}
	api := &fakePolicyAPI{snapshot: s}
	config := policyConfig(false, `{ explicit = "dev" }`)
	resource.Test(t, resource.TestCase{
		IsUnitTest: true, ProtoV6ProviderFactories: policyFactories(api),
		CheckDestroy: checkCalls(api, 1),
		Steps: []resource.TestStep{
			{Config: config, Check: checkCalls(api, 0)},
			{
				PreConfig: func() {
					api.mu.Lock()
					defer api.mu.Unlock()
					start := api.snapshotCount
					api.beforeSnapshot = func(f *fakePolicyAPI) error {
						if f.snapshotCount == start+2 {
							f.move("id-explicit", "prod")
						}
						return nil
					}
				},
				Config:      config,
				ExpectError: regexp.MustCompile("assignments changed during planning"),
			},
			{
				PreConfig: func() {
					api.mu.Lock()
					defer api.mu.Unlock()
					api.beforeSnapshot = nil
					if len(api.calls) != 0 {
						t.Fatal("rejected late drift caused an unplanned write")
					}
				},
				Config: config,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("axm_device_assignment_policy.test", plancheck.ResourceActionUpdate),
				}},
				Check: resource.ComposeTestCheckFunc(checkCalls(api, 1), resource.TestCheckResourceAttr("axm_device_assignment_policy.test", "observed_assignments.explicit", "dev")),
			},
		},
	})
}

func TestResourceTerraformImportHasNoWrites(t *testing.T) {
	s := testSnapshot()
	s.Devices = []Device{testDevice("explicit", "dev")}
	api := &fakePolicyAPI{snapshot: s}
	resource.Test(t, resource.TestCase{
		IsUnitTest: true, ProtoV6ProviderFactories: policyFactories(api),
		CheckDestroy: checkCalls(api, 0),
		Steps: []resource.TestStep{
			{Config: policyConfig(true, `{ explicit = "dev" }`), ResourceName: "axm_device_assignment_policy.test", ImportState: true, ImportStateId: policyID, ImportStatePersist: true, Check: checkCalls(api, 0)},
			{Config: policyConfig(true, `{ explicit = "dev" }`), Check: checkCalls(api, 0)},
		},
	})
}

func TestResourceTerraformImportCannotEnableWritesDuringAdoption(t *testing.T) {
	s := testSnapshot()
	s.Devices = []Device{testDevice("explicit", "prod")}
	api := &fakePolicyAPI{snapshot: s}
	resource.Test(t, resource.TestCase{
		IsUnitTest: true, ProtoV6ProviderFactories: policyFactories(api),
		CheckDestroy: checkCalls(api, 0),
		Steps: []resource.TestStep{
			{Config: policyConfig(false, `{ explicit = "dev" }`), ResourceName: "axm_device_assignment_policy.test", ImportState: true, ImportStateId: policyID, ImportStatePersist: true},
			{Config: policyConfig(false, `{ explicit = "dev" }`), ExpectError: regexp.MustCompile("adoption requires existing assignments")},
		},
	})
	if api.callCount() != 0 {
		t.Fatal("import adoption attempted assignment writes")
	}
}

func TestResourceTerraformImportBlockAdoptionHasNoWrites(t *testing.T) {
	s := testSnapshot()
	s.Devices = []Device{testDevice("explicit", "dev")}
	api := &fakePolicyAPI{snapshot: s}
	config := policyConfig(true, `{ explicit = "dev" }`) + `
import {
  to = axm_device_assignment_policy.test
  id = "device-assignments"
}
`
	resource.Test(t, resource.TestCase{
		IsUnitTest: true, ProtoV6ProviderFactories: policyFactories(api),
		CheckDestroy: checkCalls(api, 0),
		Steps: []resource.TestStep{
			{Config: config, Check: resource.ComposeTestCheckFunc(checkCalls(api, 0), resource.TestCheckResourceAttr("axm_device_assignment_policy.test", "observed_assignments.explicit", "dev"))},
			{Config: config, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}}, Check: checkCalls(api, 0)},
		},
	})
}

func TestResourceTerraformImportBlockRejectsMismatchedAdoption(t *testing.T) {
	s := testSnapshot()
	s.Devices = []Device{testDevice("explicit", "prod")}
	api := &fakePolicyAPI{snapshot: s}
	config := policyConfig(false, `{ explicit = "dev" }`) + `
import {
  to = axm_device_assignment_policy.test
  id = "device-assignments"
}
`
	resource.Test(t, resource.TestCase{
		IsUnitTest: true, ProtoV6ProviderFactories: policyFactories(api),
		CheckDestroy: checkCalls(api, 0),
		Steps:        []resource.TestStep{{Config: config, ExpectError: regexp.MustCompile("adoption requires existing assignments")}},
	})
	if api.callCount() != 0 {
		t.Fatal("import block attempted assignment writes")
	}
}

func TestResourceTerraformRejectsChangesDuringCreate(t *testing.T) {
	s := testSnapshot()
	s.Devices = []Device{testDevice("explicit", "prod")}
	api := &fakePolicyAPI{snapshot: s}
	resource.Test(t, resource.TestCase{
		IsUnitTest: true, ProtoV6ProviderFactories: policyFactories(api),
		ErrorCheck: func(err error) error {
			if err == nil || !strings.Contains(err.Error(), "adoption requires existing assignments") {
				return fmt.Errorf("expected adoption rejection, got %v", err)
			}
			if api.callCount() != 0 {
				return errors.New("create attempted an assignment write")
			}
			return nil
		},
		Steps: []resource.TestStep{{Config: policyConfig(false, `{ explicit = "dev" }`)}},
	})
}
