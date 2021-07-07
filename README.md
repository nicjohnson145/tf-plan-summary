# tf-plan-summary
Generate condensed summaries of terraform plans

## Usage

Pipe the json representation of a terraform plan into the tool to have the summarized version output
to stdout

```
terraform plan -out=plan.tfplan
terraform show --json plan.tfplan | tf-plan-summary
```

"Read" operations can be excluded from the output via the `-x/-exclude-reads` options
