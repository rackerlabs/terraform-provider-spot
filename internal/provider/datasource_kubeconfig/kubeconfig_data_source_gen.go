// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package datasource_kubeconfig

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func KubeconfigDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				Description:         "ID of the cloudspace",
				MarkdownDescription: "ID of the cloudspace",
			},
			"kubeconfigs": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cluster": schema.StringAttribute{
							Computed:            true,
							Description:         "Name of the cluster",
							MarkdownDescription: "Name of the cluster",
						},
						"exec": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"api_version": schema.StringAttribute{
									Computed:            true,
									Description:         "API version",
									MarkdownDescription: "API version",
								},
								"args": schema.ListAttribute{
									ElementType: types.StringType,
									Computed:    true,
								},
								"command": schema.StringAttribute{
									Computed:            true,
									Description:         "Command to execute",
									MarkdownDescription: "Command to execute",
								},
								"env": schema.MapAttribute{
									ElementType: types.StringType,
									Computed:    true,
								},
							},
							CustomType: ExecType{
								ObjectType: types.ObjectType{
									AttrTypes: ExecValue{}.AttributeTypes(ctx),
								},
							},
							Computed: true,
						},
						"host": schema.StringAttribute{
							Computed:            true,
							Description:         "Kube api server api endpoint",
							MarkdownDescription: "Kube api server api endpoint",
						},
						"insecure": schema.BoolAttribute{
							Computed:            true,
							Description:         "Insecure flag",
							MarkdownDescription: "Insecure flag",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							Description:         "Name of the kubeconfig context",
							MarkdownDescription: "Name of the kubeconfig context",
						},
						"token": schema.StringAttribute{
							Computed:            true,
							Sensitive:           true,
							Description:         "Token of your service account",
							MarkdownDescription: "Token of your service account",
						},
						"username": schema.StringAttribute{
							Computed:            true,
							Description:         "Name of the user",
							MarkdownDescription: "Name of the user",
						},
					},
					CustomType: KubeconfigsType{
						ObjectType: types.ObjectType{
							AttrTypes: KubeconfigsValue{}.AttributeTypes(ctx),
						},
					},
				},
				Computed: true,
			},
			"raw": schema.StringAttribute{
				Computed:            true,
				Description:         "Kubeconfig blob",
				MarkdownDescription: "Kubeconfig blob",
			},
		},
	}
}

type KubeconfigModel struct {
	Id          types.String `tfsdk:"id"`
	Kubeconfigs types.List   `tfsdk:"kubeconfigs"`
	Raw         types.String `tfsdk:"raw"`
}

var _ basetypes.ObjectTypable = KubeconfigsType{}

type KubeconfigsType struct {
	basetypes.ObjectType
}

func (t KubeconfigsType) Equal(o attr.Type) bool {
	other, ok := o.(KubeconfigsType)

	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t KubeconfigsType) String() string {
	return "KubeconfigsType"
}

func (t KubeconfigsType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributes := in.Attributes()

	clusterAttribute, ok := attributes["cluster"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`cluster is missing from object`)

		return nil, diags
	}

	clusterVal, ok := clusterAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`cluster expected to be basetypes.StringValue, was: %T`, clusterAttribute))
	}

	execAttribute, ok := attributes["exec"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`exec is missing from object`)

		return nil, diags
	}

	execVal, ok := execAttribute.(basetypes.ObjectValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`exec expected to be basetypes.ObjectValue, was: %T`, execAttribute))
	}

	hostAttribute, ok := attributes["host"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`host is missing from object`)

		return nil, diags
	}

	hostVal, ok := hostAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`host expected to be basetypes.StringValue, was: %T`, hostAttribute))
	}

	insecureAttribute, ok := attributes["insecure"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`insecure is missing from object`)

		return nil, diags
	}

	insecureVal, ok := insecureAttribute.(basetypes.BoolValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`insecure expected to be basetypes.BoolValue, was: %T`, insecureAttribute))
	}

	nameAttribute, ok := attributes["name"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`name is missing from object`)

		return nil, diags
	}

	nameVal, ok := nameAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`name expected to be basetypes.StringValue, was: %T`, nameAttribute))
	}

	tokenAttribute, ok := attributes["token"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`token is missing from object`)

		return nil, diags
	}

	tokenVal, ok := tokenAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`token expected to be basetypes.StringValue, was: %T`, tokenAttribute))
	}

	usernameAttribute, ok := attributes["username"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`username is missing from object`)

		return nil, diags
	}

	usernameVal, ok := usernameAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`username expected to be basetypes.StringValue, was: %T`, usernameAttribute))
	}

	if diags.HasError() {
		return nil, diags
	}

	return KubeconfigsValue{
		Cluster:  clusterVal,
		Exec:     execVal,
		Host:     hostVal,
		Insecure: insecureVal,
		Name:     nameVal,
		Token:    tokenVal,
		Username: usernameVal,
		state:    attr.ValueStateKnown,
	}, diags
}

func NewKubeconfigsValueNull() KubeconfigsValue {
	return KubeconfigsValue{
		state: attr.ValueStateNull,
	}
}

func NewKubeconfigsValueUnknown() KubeconfigsValue {
	return KubeconfigsValue{
		state: attr.ValueStateUnknown,
	}
}

func NewKubeconfigsValue(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) (KubeconfigsValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/521
	ctx := context.Background()

	for name, attributeType := range attributeTypes {
		attribute, ok := attributes[name]

		if !ok {
			diags.AddError(
				"Missing KubeconfigsValue Attribute Value",
				"While creating a KubeconfigsValue value, a missing attribute value was detected. "+
					"A KubeconfigsValue must contain values for all attributes, even if null or unknown. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("KubeconfigsValue Attribute Name (%s) Expected Type: %s", name, attributeType.String()),
			)

			continue
		}

		if !attributeType.Equal(attribute.Type(ctx)) {
			diags.AddError(
				"Invalid KubeconfigsValue Attribute Type",
				"While creating a KubeconfigsValue value, an invalid attribute value was detected. "+
					"A KubeconfigsValue must use a matching attribute type for the value. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("KubeconfigsValue Attribute Name (%s) Expected Type: %s\n", name, attributeType.String())+
					fmt.Sprintf("KubeconfigsValue Attribute Name (%s) Given Type: %s", name, attribute.Type(ctx)),
			)
		}
	}

	for name := range attributes {
		_, ok := attributeTypes[name]

		if !ok {
			diags.AddError(
				"Extra KubeconfigsValue Attribute Value",
				"While creating a KubeconfigsValue value, an extra attribute value was detected. "+
					"A KubeconfigsValue must not contain values beyond the expected attribute types. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Extra KubeconfigsValue Attribute Name: %s", name),
			)
		}
	}

	if diags.HasError() {
		return NewKubeconfigsValueUnknown(), diags
	}

	clusterAttribute, ok := attributes["cluster"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`cluster is missing from object`)

		return NewKubeconfigsValueUnknown(), diags
	}

	clusterVal, ok := clusterAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`cluster expected to be basetypes.StringValue, was: %T`, clusterAttribute))
	}

	execAttribute, ok := attributes["exec"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`exec is missing from object`)

		return NewKubeconfigsValueUnknown(), diags
	}

	execVal, ok := execAttribute.(basetypes.ObjectValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`exec expected to be basetypes.ObjectValue, was: %T`, execAttribute))
	}

	hostAttribute, ok := attributes["host"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`host is missing from object`)

		return NewKubeconfigsValueUnknown(), diags
	}

	hostVal, ok := hostAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`host expected to be basetypes.StringValue, was: %T`, hostAttribute))
	}

	insecureAttribute, ok := attributes["insecure"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`insecure is missing from object`)

		return NewKubeconfigsValueUnknown(), diags
	}

	insecureVal, ok := insecureAttribute.(basetypes.BoolValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`insecure expected to be basetypes.BoolValue, was: %T`, insecureAttribute))
	}

	nameAttribute, ok := attributes["name"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`name is missing from object`)

		return NewKubeconfigsValueUnknown(), diags
	}

	nameVal, ok := nameAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`name expected to be basetypes.StringValue, was: %T`, nameAttribute))
	}

	tokenAttribute, ok := attributes["token"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`token is missing from object`)

		return NewKubeconfigsValueUnknown(), diags
	}

	tokenVal, ok := tokenAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`token expected to be basetypes.StringValue, was: %T`, tokenAttribute))
	}

	usernameAttribute, ok := attributes["username"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`username is missing from object`)

		return NewKubeconfigsValueUnknown(), diags
	}

	usernameVal, ok := usernameAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`username expected to be basetypes.StringValue, was: %T`, usernameAttribute))
	}

	if diags.HasError() {
		return NewKubeconfigsValueUnknown(), diags
	}

	return KubeconfigsValue{
		Cluster:  clusterVal,
		Exec:     execVal,
		Host:     hostVal,
		Insecure: insecureVal,
		Name:     nameVal,
		Token:    tokenVal,
		Username: usernameVal,
		state:    attr.ValueStateKnown,
	}, diags
}

func NewKubeconfigsValueMust(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) KubeconfigsValue {
	object, diags := NewKubeconfigsValue(attributeTypes, attributes)

	if diags.HasError() {
		// This could potentially be added to the diag package.
		diagsStrings := make([]string, 0, len(diags))

		for _, diagnostic := range diags {
			diagsStrings = append(diagsStrings, fmt.Sprintf(
				"%s | %s | %s",
				diagnostic.Severity(),
				diagnostic.Summary(),
				diagnostic.Detail()))
		}

		panic("NewKubeconfigsValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}

	return object
}

func (t KubeconfigsType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.Type() == nil {
		return NewKubeconfigsValueNull(), nil
	}

	if !in.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), in.Type())
	}

	if !in.IsKnown() {
		return NewKubeconfigsValueUnknown(), nil
	}

	if in.IsNull() {
		return NewKubeconfigsValueNull(), nil
	}

	attributes := map[string]attr.Value{}

	val := map[string]tftypes.Value{}

	err := in.As(&val)

	if err != nil {
		return nil, err
	}

	for k, v := range val {
		a, err := t.AttrTypes[k].ValueFromTerraform(ctx, v)

		if err != nil {
			return nil, err
		}

		attributes[k] = a
	}

	return NewKubeconfigsValueMust(KubeconfigsValue{}.AttributeTypes(ctx), attributes), nil
}

func (t KubeconfigsType) ValueType(ctx context.Context) attr.Value {
	return KubeconfigsValue{}
}

var _ basetypes.ObjectValuable = KubeconfigsValue{}

type KubeconfigsValue struct {
	Cluster  basetypes.StringValue `tfsdk:"cluster"`
	Exec     basetypes.ObjectValue `tfsdk:"exec"`
	Host     basetypes.StringValue `tfsdk:"host"`
	Insecure basetypes.BoolValue   `tfsdk:"insecure"`
	Name     basetypes.StringValue `tfsdk:"name"`
	Token    basetypes.StringValue `tfsdk:"token"`
	Username basetypes.StringValue `tfsdk:"username"`
	state    attr.ValueState
}

func (v KubeconfigsValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	attrTypes := make(map[string]tftypes.Type, 7)

	var val tftypes.Value
	var err error

	attrTypes["cluster"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["exec"] = basetypes.ObjectType{
		AttrTypes: ExecValue{}.AttributeTypes(ctx),
	}.TerraformType(ctx)
	attrTypes["host"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["insecure"] = basetypes.BoolType{}.TerraformType(ctx)
	attrTypes["name"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["token"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["username"] = basetypes.StringType{}.TerraformType(ctx)

	objectType := tftypes.Object{AttributeTypes: attrTypes}

	switch v.state {
	case attr.ValueStateKnown:
		vals := make(map[string]tftypes.Value, 7)

		val, err = v.Cluster.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["cluster"] = val

		val, err = v.Exec.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["exec"] = val

		val, err = v.Host.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["host"] = val

		val, err = v.Insecure.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["insecure"] = val

		val, err = v.Name.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["name"] = val

		val, err = v.Token.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["token"] = val

		val, err = v.Username.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["username"] = val

		if err := tftypes.ValidateValue(objectType, vals); err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		return tftypes.NewValue(objectType, vals), nil
	case attr.ValueStateNull:
		return tftypes.NewValue(objectType, nil), nil
	case attr.ValueStateUnknown:
		return tftypes.NewValue(objectType, tftypes.UnknownValue), nil
	default:
		panic(fmt.Sprintf("unhandled Object state in ToTerraformValue: %s", v.state))
	}
}

func (v KubeconfigsValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v KubeconfigsValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v KubeconfigsValue) String() string {
	return "KubeconfigsValue"
}

func (v KubeconfigsValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	var exec basetypes.ObjectValue

	if v.Exec.IsNull() {
		exec = types.ObjectNull(
			ExecValue{}.AttributeTypes(ctx),
		)
	}

	if v.Exec.IsUnknown() {
		exec = types.ObjectUnknown(
			ExecValue{}.AttributeTypes(ctx),
		)
	}

	if !v.Exec.IsNull() && !v.Exec.IsUnknown() {
		exec = types.ObjectValueMust(
			ExecValue{}.AttributeTypes(ctx),
			v.Exec.Attributes(),
		)
	}

	objVal, diags := types.ObjectValue(
		map[string]attr.Type{
			"cluster": basetypes.StringType{},
			"exec": basetypes.ObjectType{
				AttrTypes: ExecValue{}.AttributeTypes(ctx),
			},
			"host":     basetypes.StringType{},
			"insecure": basetypes.BoolType{},
			"name":     basetypes.StringType{},
			"token":    basetypes.StringType{},
			"username": basetypes.StringType{},
		},
		map[string]attr.Value{
			"cluster":  v.Cluster,
			"exec":     exec,
			"host":     v.Host,
			"insecure": v.Insecure,
			"name":     v.Name,
			"token":    v.Token,
			"username": v.Username,
		})

	return objVal, diags
}

func (v KubeconfigsValue) Equal(o attr.Value) bool {
	other, ok := o.(KubeconfigsValue)

	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state != attr.ValueStateKnown {
		return true
	}

	if !v.Cluster.Equal(other.Cluster) {
		return false
	}

	if !v.Exec.Equal(other.Exec) {
		return false
	}

	if !v.Host.Equal(other.Host) {
		return false
	}

	if !v.Insecure.Equal(other.Insecure) {
		return false
	}

	if !v.Name.Equal(other.Name) {
		return false
	}

	if !v.Token.Equal(other.Token) {
		return false
	}

	if !v.Username.Equal(other.Username) {
		return false
	}

	return true
}

func (v KubeconfigsValue) Type(ctx context.Context) attr.Type {
	return KubeconfigsType{
		basetypes.ObjectType{
			AttrTypes: v.AttributeTypes(ctx),
		},
	}
}

func (v KubeconfigsValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"cluster": basetypes.StringType{},
		"exec": basetypes.ObjectType{
			AttrTypes: ExecValue{}.AttributeTypes(ctx),
		},
		"host":     basetypes.StringType{},
		"insecure": basetypes.BoolType{},
		"name":     basetypes.StringType{},
		"token":    basetypes.StringType{},
		"username": basetypes.StringType{},
	}
}

var _ basetypes.ObjectTypable = ExecType{}

type ExecType struct {
	basetypes.ObjectType
}

func (t ExecType) Equal(o attr.Type) bool {
	other, ok := o.(ExecType)

	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t ExecType) String() string {
	return "ExecType"
}

func (t ExecType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributes := in.Attributes()

	apiVersionAttribute, ok := attributes["api_version"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`api_version is missing from object`)

		return nil, diags
	}

	apiVersionVal, ok := apiVersionAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`api_version expected to be basetypes.StringValue, was: %T`, apiVersionAttribute))
	}

	argsAttribute, ok := attributes["args"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`args is missing from object`)

		return nil, diags
	}

	argsVal, ok := argsAttribute.(basetypes.ListValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`args expected to be basetypes.ListValue, was: %T`, argsAttribute))
	}

	commandAttribute, ok := attributes["command"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`command is missing from object`)

		return nil, diags
	}

	commandVal, ok := commandAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`command expected to be basetypes.StringValue, was: %T`, commandAttribute))
	}

	envAttribute, ok := attributes["env"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`env is missing from object`)

		return nil, diags
	}

	envVal, ok := envAttribute.(basetypes.MapValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`env expected to be basetypes.MapValue, was: %T`, envAttribute))
	}

	if diags.HasError() {
		return nil, diags
	}

	return ExecValue{
		ApiVersion: apiVersionVal,
		Args:       argsVal,
		Command:    commandVal,
		Env:        envVal,
		state:      attr.ValueStateKnown,
	}, diags
}

func NewExecValueNull() ExecValue {
	return ExecValue{
		state: attr.ValueStateNull,
	}
}

func NewExecValueUnknown() ExecValue {
	return ExecValue{
		state: attr.ValueStateUnknown,
	}
}

func NewExecValue(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) (ExecValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/521
	ctx := context.Background()

	for name, attributeType := range attributeTypes {
		attribute, ok := attributes[name]

		if !ok {
			diags.AddError(
				"Missing ExecValue Attribute Value",
				"While creating a ExecValue value, a missing attribute value was detected. "+
					"A ExecValue must contain values for all attributes, even if null or unknown. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("ExecValue Attribute Name (%s) Expected Type: %s", name, attributeType.String()),
			)

			continue
		}

		if !attributeType.Equal(attribute.Type(ctx)) {
			diags.AddError(
				"Invalid ExecValue Attribute Type",
				"While creating a ExecValue value, an invalid attribute value was detected. "+
					"A ExecValue must use a matching attribute type for the value. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("ExecValue Attribute Name (%s) Expected Type: %s\n", name, attributeType.String())+
					fmt.Sprintf("ExecValue Attribute Name (%s) Given Type: %s", name, attribute.Type(ctx)),
			)
		}
	}

	for name := range attributes {
		_, ok := attributeTypes[name]

		if !ok {
			diags.AddError(
				"Extra ExecValue Attribute Value",
				"While creating a ExecValue value, an extra attribute value was detected. "+
					"A ExecValue must not contain values beyond the expected attribute types. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Extra ExecValue Attribute Name: %s", name),
			)
		}
	}

	if diags.HasError() {
		return NewExecValueUnknown(), diags
	}

	apiVersionAttribute, ok := attributes["api_version"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`api_version is missing from object`)

		return NewExecValueUnknown(), diags
	}

	apiVersionVal, ok := apiVersionAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`api_version expected to be basetypes.StringValue, was: %T`, apiVersionAttribute))
	}

	argsAttribute, ok := attributes["args"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`args is missing from object`)

		return NewExecValueUnknown(), diags
	}

	argsVal, ok := argsAttribute.(basetypes.ListValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`args expected to be basetypes.ListValue, was: %T`, argsAttribute))
	}

	commandAttribute, ok := attributes["command"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`command is missing from object`)

		return NewExecValueUnknown(), diags
	}

	commandVal, ok := commandAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`command expected to be basetypes.StringValue, was: %T`, commandAttribute))
	}

	envAttribute, ok := attributes["env"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`env is missing from object`)

		return NewExecValueUnknown(), diags
	}

	envVal, ok := envAttribute.(basetypes.MapValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`env expected to be basetypes.MapValue, was: %T`, envAttribute))
	}

	if diags.HasError() {
		return NewExecValueUnknown(), diags
	}

	return ExecValue{
		ApiVersion: apiVersionVal,
		Args:       argsVal,
		Command:    commandVal,
		Env:        envVal,
		state:      attr.ValueStateKnown,
	}, diags
}

func NewExecValueMust(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) ExecValue {
	object, diags := NewExecValue(attributeTypes, attributes)

	if diags.HasError() {
		// This could potentially be added to the diag package.
		diagsStrings := make([]string, 0, len(diags))

		for _, diagnostic := range diags {
			diagsStrings = append(diagsStrings, fmt.Sprintf(
				"%s | %s | %s",
				diagnostic.Severity(),
				diagnostic.Summary(),
				diagnostic.Detail()))
		}

		panic("NewExecValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}

	return object
}

func (t ExecType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.Type() == nil {
		return NewExecValueNull(), nil
	}

	if !in.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), in.Type())
	}

	if !in.IsKnown() {
		return NewExecValueUnknown(), nil
	}

	if in.IsNull() {
		return NewExecValueNull(), nil
	}

	attributes := map[string]attr.Value{}

	val := map[string]tftypes.Value{}

	err := in.As(&val)

	if err != nil {
		return nil, err
	}

	for k, v := range val {
		a, err := t.AttrTypes[k].ValueFromTerraform(ctx, v)

		if err != nil {
			return nil, err
		}

		attributes[k] = a
	}

	return NewExecValueMust(ExecValue{}.AttributeTypes(ctx), attributes), nil
}

func (t ExecType) ValueType(ctx context.Context) attr.Value {
	return ExecValue{}
}

var _ basetypes.ObjectValuable = ExecValue{}

type ExecValue struct {
	ApiVersion basetypes.StringValue `tfsdk:"api_version"`
	Args       basetypes.ListValue   `tfsdk:"args"`
	Command    basetypes.StringValue `tfsdk:"command"`
	Env        basetypes.MapValue    `tfsdk:"env"`
	state      attr.ValueState
}

func (v ExecValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	attrTypes := make(map[string]tftypes.Type, 4)

	var val tftypes.Value
	var err error

	attrTypes["api_version"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["args"] = basetypes.ListType{
		ElemType: types.StringType,
	}.TerraformType(ctx)
	attrTypes["command"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["env"] = basetypes.MapType{
		ElemType: types.StringType,
	}.TerraformType(ctx)

	objectType := tftypes.Object{AttributeTypes: attrTypes}

	switch v.state {
	case attr.ValueStateKnown:
		vals := make(map[string]tftypes.Value, 4)

		val, err = v.ApiVersion.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["api_version"] = val

		val, err = v.Args.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["args"] = val

		val, err = v.Command.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["command"] = val

		val, err = v.Env.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["env"] = val

		if err := tftypes.ValidateValue(objectType, vals); err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		return tftypes.NewValue(objectType, vals), nil
	case attr.ValueStateNull:
		return tftypes.NewValue(objectType, nil), nil
	case attr.ValueStateUnknown:
		return tftypes.NewValue(objectType, tftypes.UnknownValue), nil
	default:
		panic(fmt.Sprintf("unhandled Object state in ToTerraformValue: %s", v.state))
	}
}

func (v ExecValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v ExecValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v ExecValue) String() string {
	return "ExecValue"
}

func (v ExecValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	argsVal, d := types.ListValue(types.StringType, v.Args.Elements())

	diags.Append(d...)

	if d.HasError() {
		return types.ObjectUnknown(map[string]attr.Type{
			"api_version": basetypes.StringType{},
			"args": basetypes.ListType{
				ElemType: types.StringType,
			},
			"command": basetypes.StringType{},
			"env": basetypes.MapType{
				ElemType: types.StringType,
			},
		}), diags
	}

	envVal, d := types.MapValue(types.StringType, v.Env.Elements())

	diags.Append(d...)

	if d.HasError() {
		return types.ObjectUnknown(map[string]attr.Type{
			"api_version": basetypes.StringType{},
			"args": basetypes.ListType{
				ElemType: types.StringType,
			},
			"command": basetypes.StringType{},
			"env": basetypes.MapType{
				ElemType: types.StringType,
			},
		}), diags
	}

	objVal, diags := types.ObjectValue(
		map[string]attr.Type{
			"api_version": basetypes.StringType{},
			"args": basetypes.ListType{
				ElemType: types.StringType,
			},
			"command": basetypes.StringType{},
			"env": basetypes.MapType{
				ElemType: types.StringType,
			},
		},
		map[string]attr.Value{
			"api_version": v.ApiVersion,
			"args":        argsVal,
			"command":     v.Command,
			"env":         envVal,
		})

	return objVal, diags
}

func (v ExecValue) Equal(o attr.Value) bool {
	other, ok := o.(ExecValue)

	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state != attr.ValueStateKnown {
		return true
	}

	if !v.ApiVersion.Equal(other.ApiVersion) {
		return false
	}

	if !v.Args.Equal(other.Args) {
		return false
	}

	if !v.Command.Equal(other.Command) {
		return false
	}

	if !v.Env.Equal(other.Env) {
		return false
	}

	return true
}

func (v ExecValue) Type(ctx context.Context) attr.Type {
	return ExecType{
		basetypes.ObjectType{
			AttrTypes: v.AttributeTypes(ctx),
		},
	}
}

func (v ExecValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"api_version": basetypes.StringType{},
		"args": basetypes.ListType{
			ElemType: types.StringType,
		},
		"command": basetypes.StringType{},
		"env": basetypes.MapType{
			ElemType: types.StringType,
		},
	}
}
