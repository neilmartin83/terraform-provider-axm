// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package device_assignment_policy

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"reflect"
	"sort"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/neilmartin83/terraform-provider-axm/internal/client"
	"github.com/neilmartin83/terraform-provider-axm/internal/common"
)

const policyID = "device-assignments"
const approvalKey = "assignment_approval_v1"
const operationTimeout = 30 * time.Minute

var _ resource.Resource = &AssignmentPolicyResource{}
var _ resource.ResourceWithConfigure = &AssignmentPolicyResource{}
var _ resource.ResourceWithValidateConfig = &AssignmentPolicyResource{}
var _ resource.ResourceWithModifyPlan = &AssignmentPolicyResource{}
var _ resource.ResourceWithImportState = &AssignmentPolicyResource{}

type policyAPI interface {
	Snapshot(context.Context) (Snapshot, error)
	Device(context.Context, string) (Device, error)
	Server(context.Context, string) (Server, error)
	Assign(context.Context, string, string) error
}

// AssignmentPolicyResource is the sole owner of assignment transitions in an account.
type AssignmentPolicyResource struct{ api policyAPI }

type policyModel struct {
	ID                     types.String `tfsdk:"id"`
	Assignments            types.Map    `tfsdk:"assignments"`
	AuthoritativeServerIDs types.Set    `tfsdk:"authoritative_server_ids"`
	FallbackServerID       types.String `tfsdk:"fallback_server_id"`
	ServerNames            types.Map    `tfsdk:"server_names"`
	AdoptionOnly           types.Bool   `tfsdk:"adoption_only"`
	MaxChanges             types.Int64  `tfsdk:"max_changes"`
	ObservedAssignments    types.Map    `tfsdk:"observed_assignments"`
	DeviceCount            types.Int64  `tfsdk:"device_count"`
}

type approval struct {
	Version    int
	PolicyHash string
	Plan       Plan
}

// NewAssignmentPolicyResource returns the assignment-only policy resource.
func NewAssignmentPolicyResource() resource.Resource { return &AssignmentPolicyResource{} }

func (r *AssignmentPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_assignment_policy"
}

func (r *AssignmentPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Enforces explicit device destinations and moves unlisted members of authoritative MDM servers to a fallback MDM server. Use only one policy per Apple account and keep device_management_service assignment management disabled. Create/import are read-only adoption; delete only forgets the policy. All identifiers remain in state and saved plans even when marked sensitive.\n\n" +
			"Explicit assignments take precedence, including UNASSIGNED and assignments to the fallback server. Devices outside explicit assignments and authoritative servers remain unmanaged. Removing an explicit entry while its device remains on an authoritative server makes it a fallback candidate. Apple-managed and Configurator destinations are not supported.\n\n" +
			"Start with adoption_only enabled and a configuration matching existing assignments. Once adoption succeeds, review a separate plan to enable enforcement. Corrections happen on Terraform apply, not continuously. Plans retain the exact observed assignments and intended transitions; changed observations require a fresh plan. Reads and writes have a 30-minute operation deadline.\n\n" +
			"Assignment writes are not retried automatically. A failed or uncertain activity stops subsequent writes without rollback; inspect current assignments and create a fresh plan before continuing. Immediate device and destination checks reduce races but cannot atomically lock the Apple console or other API writers. Coordinate those writers during an apply, and use a small max_changes limit.",
		Attributes: map[string]schema.Attribute{
			"id":                       schema.StringAttribute{Computed: true, Description: "Account-local singleton policy identifier."},
			"assignments":              schema.MapAttribute{Required: true, Sensitive: true, ElementType: types.StringType, Description: "Hardware serial number to pinned MDM server ID, or UNASSIGNED. Explicit intent overrides authoritative-server fallback."},
			"authoritative_server_ids": schema.SetAttribute{Required: true, ElementType: types.StringType, Description: "MDM servers whose unlisted members must move to the fallback server."},
			"fallback_server_id":       schema.StringAttribute{Required: true, Description: "Pinned MDM destination for unlisted authoritative-server members. Its own complete membership stays unmanaged."},
			"server_names":             schema.MapAttribute{Required: true, ElementType: types.StringType, Description: "Expected names keyed by server ID. Every explicit destination, authoritative server and fallback must have a matching live MDM identity."},
			"adoption_only":            schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true), Description: "Defaults to true: refuse any assignment changes. Set false only after reviewing adoption and an enforcement plan."},
			"max_changes":              schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(1), Description: "Maximum device transitions in one apply, checked again before writes. Zero permits no transitions."},
			"observed_assignments":     schema.MapAttribute{Computed: true, Sensitive: true, ElementType: types.StringType, Description: "Observed explicit-device and authoritative-server membership. The plan shows the resolved final assignments; extra members disappear from this map after fallback."},
			"device_count":             schema.Int64Attribute{Computed: true, Description: "Number of explicit device entries; excludes transient fallback candidates."},
		},
	}
}

func (r *AssignmentPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	c, diagnostics := common.ConfigureClient(req.ProviderData, "Resource")
	resp.Diagnostics.Append(diagnostics...)
	if c != nil {
		r.api = clientAdapter{c}
	}
}

func (m policyModel) policy(ctx context.Context) (Policy, error) {
	if m.Assignments.IsNull() || m.Assignments.IsUnknown() || m.AuthoritativeServerIDs.IsNull() || m.AuthoritativeServerIDs.IsUnknown() || m.FallbackServerID.IsNull() || m.FallbackServerID.IsUnknown() || m.ServerNames.IsNull() || m.ServerNames.IsUnknown() || m.AdoptionOnly.IsUnknown() || m.MaxChanges.IsUnknown() {
		return Policy{}, errors.New("all assignment policy inputs must be known before planning")
	}
	p := Policy{FallbackServerID: m.FallbackServerID.ValueString(), AdoptionOnly: true, MaxChanges: 1}
	if !m.AdoptionOnly.IsNull() {
		p.AdoptionOnly = m.AdoptionOnly.ValueBool()
	}
	if !m.MaxChanges.IsNull() {
		p.MaxChanges = m.MaxChanges.ValueInt64()
	}
	if m.Assignments.ElementsAs(ctx, &p.Assignments, false).HasError() || m.AuthoritativeServerIDs.ElementsAs(ctx, &p.AuthoritativeServerIDs, false).HasError() || m.ServerNames.ElementsAs(ctx, &p.ServerNames, false).HasError() {
		return Policy{}, errors.New("assignment policy inputs must contain known, non-null strings")
	}
	sort.Strings(p.AuthoritativeServerIDs)
	return p, ValidatePolicy(p)
}

func fail(d *diag.Diagnostics, err error) {
	if err != nil {
		d.AddError("Assignment policy rejected", err.Error())
	}
}

func (r *AssignmentPolicyResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var m policyModel
	if req.Config.Get(ctx, &m).HasError() {
		fail(&resp.Diagnostics, errors.New("assignment policy configuration is invalid"))
		return
	}
	if m.Assignments.IsUnknown() || m.AuthoritativeServerIDs.IsUnknown() || m.FallbackServerID.IsUnknown() || m.ServerNames.IsUnknown() || m.AdoptionOnly.IsUnknown() || m.MaxChanges.IsUnknown() {
		return
	}
	for _, values := range []types.Map{m.Assignments, m.ServerNames} {
		for _, value := range values.Elements() {
			if value.IsUnknown() {
				return
			}
		}
	}
	for _, value := range m.AuthoritativeServerIDs.Elements() {
		if value.IsUnknown() {
			return
		}
	}
	_, err := m.policy(ctx)
	fail(&resp.Diagnostics, err)
}

func (r *AssignmentPolicyResource) prepare(ctx context.Context, p Policy) (Plan, error) {
	s, err := r.snapshot(ctx)
	if err != nil {
		return Plan{}, err
	}
	return BuildPlan(p, s)
}

func (r *AssignmentPolicyResource) snapshot(ctx context.Context) (Snapshot, error) {
	if r.api == nil {
		return Snapshot{}, errors.New("assignment policy API client is unavailable")
	}
	s, err := r.api.Snapshot(ctx)
	if err != nil {
		return Snapshot{}, errors.New("could not read complete assignment inventory; no writes attempted")
	}
	return s, nil
}

func checkLimit(p Policy, plan Plan, adopting bool) error {
	if len(plan.Changes) > 0 && (p.AdoptionOnly || adopting) {
		return errors.New("adoption requires existing assignments to match the policy; no writes permitted")
	}
	if int64(len(plan.Changes)) > p.MaxChanges {
		return errors.New("planned device transitions exceed max_changes")
	}
	return nil
}

func policyHash(p Policy) string {
	b, _ := json.Marshal(p)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func (r *AssignmentPolicyResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, operationTimeout)
	defer cancel()
	var m policyModel
	if req.Plan.Get(ctx, &m).HasError() {
		fail(&resp.Diagnostics, errors.New("assignment plan is invalid"))
		return
	}
	p, err := m.policy(ctx)
	if err != nil {
		fail(&resp.Diagnostics, err)
		return
	}
	snapshot, err := r.snapshot(ctx)
	if err != nil {
		fail(&resp.Diagnostics, err)
		return
	}
	plan, err := BuildPlan(p, snapshot)
	if err == nil {
		adopting := req.State.Raw.IsNull()
		if !adopting {
			var previous policyModel
			if req.State.Get(ctx, &previous).HasError() {
				fail(&resp.Diagnostics, errors.New("prior assignment state is invalid"))
				return
			}
			adopting = previous.Assignments.IsNull()
			if !adopting {
				previousPolicy, previousErr := previous.policy(ctx)
				if previousErr != nil {
					fail(&resp.Diagnostics, previousErr)
					return
				}
				refreshed, previousErr := BuildPlan(previousPolicy, snapshot)
				if previousErr != nil {
					fail(&resp.Diagnostics, previousErr)
					return
				}
				var observed map[string]string
				if previous.ObservedAssignments.ElementsAs(ctx, &observed, false).HasError() || !reflect.DeepEqual(observed, refreshed.Observed) {
					fail(&resp.Diagnostics, errors.New("assignments changed during planning; create and review a fresh plan"))
					return
				}
			}
		}
		err = checkLimit(p, plan, adopting)
	}
	if err != nil {
		fail(&resp.Diagnostics, err)
		return
	}
	approved, err := json.Marshal(approval{Version: 1, PolicyHash: policyHash(p), Plan: plan})
	if err != nil || resp.Private == nil {
		fail(&resp.Diagnostics, errors.New("could not save the approved assignment plan"))
		return
	}
	resp.Diagnostics.Append(resp.Private.SetKey(ctx, approvalKey, approved)...)
	m.ID = types.StringValue(policyID)
	m.ObservedAssignments, _ = types.MapValueFrom(ctx, types.StringType, p.Assignments)
	m.DeviceCount = types.Int64Value(int64(len(p.Assignments)))
	resp.Diagnostics.Append(resp.Plan.Set(ctx, m)...)
}

func (r *AssignmentPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ctx, cancel := context.WithTimeout(ctx, operationTimeout)
	defer cancel()
	var m policyModel
	if req.State.Get(ctx, &m).HasError() {
		fail(&resp.Diagnostics, errors.New("assignment state is invalid"))
		return
	}
	if m.Assignments.IsNull() && m.ServerNames.IsNull() {
		return
	}
	p, err := m.policy(ctx)
	if err != nil {
		fail(&resp.Diagnostics, err)
		return
	}
	plan, err := r.prepare(ctx, p)
	if err != nil {
		fail(&resp.Diagnostics, err)
		return
	}
	m.ObservedAssignments, _ = types.MapValueFrom(ctx, types.StringType, plan.Observed)
	m.DeviceCount = types.Int64Value(int64(len(p.Assignments)))
	resp.Diagnostics.Append(resp.State.Set(ctx, m)...)
}

func (r *AssignmentPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	ctx, cancel := context.WithTimeout(ctx, operationTimeout)
	defer cancel()
	var m policyModel
	if req.Plan.Get(ctx, &m).HasError() {
		fail(&resp.Diagnostics, errors.New("assignment plan is invalid"))
		return
	}
	p, err := m.policy(ctx)
	if err != nil {
		fail(&resp.Diagnostics, err)
		return
	}
	plan, err := r.prepare(ctx, p)
	if err == nil {
		err = checkLimit(p, plan, true)
	}
	if err != nil {
		fail(&resp.Diagnostics, err)
		return
	}
	m.ID = types.StringValue(policyID)
	m.ObservedAssignments, _ = types.MapValueFrom(ctx, types.StringType, plan.Observed)
	m.DeviceCount = types.Int64Value(int64(len(p.Assignments)))
	resp.Diagnostics.Append(resp.State.Set(ctx, m)...)
}

func (r *AssignmentPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	ctx, cancel := context.WithTimeout(ctx, operationTimeout)
	defer cancel()
	resp.State = req.State
	var m policyModel
	if req.Plan.Get(ctx, &m).HasError() {
		fail(&resp.Diagnostics, errors.New("assignment plan is invalid"))
		return
	}
	p, err := m.policy(ctx)
	if err != nil {
		fail(&resp.Diagnostics, err)
		return
	}
	if req.Private == nil {
		fail(&resp.Diagnostics, errors.New("approved assignment plan is missing"))
		return
	}
	b, diagnostics := req.Private.GetKey(ctx, approvalKey)
	if diagnostics.HasError() {
		fail(&resp.Diagnostics, errors.New("approved assignment plan could not be read"))
		return
	}
	var approved approval
	if json.Unmarshal(b, &approved) != nil || approved.Version != 1 || approved.PolicyHash != policyHash(p) {
		fail(&resp.Diagnostics, errors.New("approved assignment plan does not match the configured policy"))
		return
	}
	var previous policyModel
	if req.State.Get(ctx, &previous).HasError() {
		fail(&resp.Diagnostics, errors.New("prior assignment state is invalid"))
		return
	}
	if err := checkLimit(p, approved.Plan, previous.Assignments.IsNull()); err != nil {
		fail(&resp.Diagnostics, err)
		return
	}
	if err := r.apply(ctx, p, approved.Plan); err != nil {
		fail(&resp.Diagnostics, err)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, m)...)
	if resp.Private != nil {
		resp.Diagnostics.Append(resp.Private.SetKey(ctx, approvalKey, nil)...)
	}
}

func (r *AssignmentPolicyResource) apply(ctx context.Context, p Policy, approved Plan) error {
	if err := checkLimit(p, approved, false); err != nil {
		return err
	}
	current, err := r.prepare(ctx, p)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(current, approved) {
		return errors.New("assignments changed after planning; create and review a fresh plan")
	}
	for _, change := range approved.Changes {
		device, err := r.api.Device(ctx, change.DeviceID)
		if err != nil {
			return errors.New("device revalidation failed; stopped before the next write")
		}
		from := device.ServerID
		if (device.Status == "ASSIGNED" && (from == "" || from == Unassigned)) || (device.Status == Unassigned && from != "") {
			return errors.New("device assignment status is inconsistent; stopped before the next write")
		}
		if device.Status == Unassigned && from == "" {
			from = Unassigned
		}
		if device.ID != change.DeviceID || device.SerialNumber != change.SerialNumber || from != change.From || (device.Status != "ASSIGNED" && device.Status != Unassigned) {
			return errors.New("device changed after planning; stopped before the next write")
		}
		if change.To != Unassigned {
			server, err := r.api.Server(ctx, change.To)
			if err != nil || server.ID != change.To || server.Type != "MDM" || server.Name != p.ServerNames[change.To] {
				return errors.New("destination identity changed after planning; stopped before the next write")
			}
		}
		destination := change.To
		if destination == Unassigned {
			destination = ""
		}
		if err := r.api.Assign(ctx, change.DeviceID, destination); err != nil {
			return errors.New("assignment did not complete with verified success; stopped without retry or rollback; inspect current assignments before continuing")
		}
	}
	final, err := r.prepare(ctx, p)
	if err != nil || len(final.Changes) != 0 {
		return errors.New("post-apply verification failed; inspect current assignments before continuing")
	}
	return nil
}

func (r *AssignmentPolicyResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func (r *AssignmentPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != policyID {
		fail(&resp.Diagnostics, errors.New("import identifier must be device-assignments; import never changes assignments"))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), policyID)...)
}

type clientAdapter struct{ c *client.Client }

func (a clientAdapter) Snapshot(ctx context.Context) (Snapshot, error) {
	s, err := a.c.GetAssignmentPolicySnapshot(ctx)
	if err != nil {
		return Snapshot{}, err
	}
	out := Snapshot{Servers: make([]Server, 0, len(s.Servers)), Devices: make([]Device, 0, len(s.Devices))}
	for _, s := range s.Servers {
		out.Servers = append(out.Servers, Server{ID: s.ID, Name: s.Name, Type: s.Type})
	}
	for _, d := range s.Devices {
		out.Devices = append(out.Devices, Device{ID: d.ID, SerialNumber: d.SerialNumber, Status: d.Status, ServerID: d.ServerID})
	}
	return out, nil
}

func (a clientAdapter) Device(ctx context.Context, id string) (Device, error) {
	d, err := a.c.GetAssignmentPolicyDevice(ctx, id)
	if err != nil {
		return Device{}, err
	}
	return Device{ID: d.ID, SerialNumber: d.SerialNumber, Status: d.Status, ServerID: d.ServerID}, nil
}

func (a clientAdapter) Server(ctx context.Context, id string) (Server, error) {
	s, err := a.c.GetAssignmentPolicyServer(ctx, id)
	if err != nil {
		return Server{}, err
	}
	return Server{ID: s.ID, Name: s.Name, Type: s.Type}, nil
}

func (a clientAdapter) Assign(ctx context.Context, id, destination string) error {
	return a.c.AssignPolicyDevice(ctx, id, destination)
}
