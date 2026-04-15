data "axm_app" "example" {
  id = "361309726"
}

output "example_app" {
  value = data.axm_app.example
}
