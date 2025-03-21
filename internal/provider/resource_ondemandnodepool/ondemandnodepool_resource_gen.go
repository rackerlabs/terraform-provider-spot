// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package resource_ondemandnodepool

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func OndemandnodepoolResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"annotations": schema.MapAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Description:         "Annotations to be applied to the nodes of the node pool",
				MarkdownDescription: "Annotations to be applied to the nodes of the node pool",
			},
			"cloudspace_name": schema.StringAttribute{
				Required:            true,
				Description:         "The name of the cloudspace.",
				MarkdownDescription: "The name of the cloudspace.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 63),
					stringvalidator.RegexMatches(regexp.MustCompile(`^[a-zA-Z0-9]([-a-zA-Z0-9]*[a-zA-Z0-9])?$`), "Must be valid kubernetes name"),
				},
			},
			"desired_server_count": schema.Int64Attribute{
				Required:            true,
				Description:         "The desired number of servers in the node pool.",
				MarkdownDescription: "The desired number of servers in the node pool.",
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"labels": schema.MapAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Description:         "Labels to be applied to the nodes of the node pool",
				MarkdownDescription: "Labels to be applied to the nodes of the node pool",
			},
			"last_updated": schema.StringAttribute{
				Computed:            true,
				Description:         "The last time the ondemandnodepool was updated.",
				MarkdownDescription: "The last time the ondemandnodepool was updated.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Computed:            true,
				Description:         "The name of the ondemandnodepool.",
				MarkdownDescription: "The name of the ondemandnodepool.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"reserved_count": schema.Int64Attribute{
				Computed:            true,
				Description:         "Number of reserved on-demand nodes.",
				MarkdownDescription: "Number of reserved on-demand nodes.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"reserved_status": schema.StringAttribute{
				Computed:            true,
				Description:         "Status of the ondemandnodepool.",
				MarkdownDescription: "Status of the ondemandnodepool.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"server_class": schema.StringAttribute{
				Required:            true,
				Description:         "The server class to be used for the node pool can be obtained from the serverclasses data source.",
				MarkdownDescription: "The server class to be used for the node pool can be obtained from the serverclasses data source.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"taints": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"effect": schema.StringAttribute{
							Required:            true,
							Description:         "The taint effect (NoSchedule, PreferNoSchedule, or NoExecute)",
							MarkdownDescription: "The taint effect (NoSchedule, PreferNoSchedule, or NoExecute)",
							Validators: []validator.String{
								stringvalidator.OneOf("NoSchedule", "PreferNoSchedule", "NoExecute"),
							},
						},
						"key": schema.StringAttribute{
							Required:            true,
							Description:         "The taint key to be applied",
							MarkdownDescription: "The taint key to be applied",
						},
						"value": schema.StringAttribute{
							Optional:            true,
							Description:         "The taint value",
							MarkdownDescription: "The taint value",
						},
					},
					CustomType: TaintsType{
						ObjectType: types.ObjectType{
							AttrTypes: TaintsValue{}.AttributeTypes(ctx),
						},
					},
				},
				Optional:            true,
				Description:         "Kubernetes taints to be applied to the nodes of the node pool",
				MarkdownDescription: "Kubernetes taints to be applied to the nodes of the node pool",
			},
		},
	}
}

type OndemandnodepoolModel struct {
	Annotations        types.Map    `tfsdk:"annotations"`
	CloudspaceName     types.String `tfsdk:"cloudspace_name"`
	DesiredServerCount types.Int64  `tfsdk:"desired_server_count"`
	Labels             types.Map    `tfsdk:"labels"`
	LastUpdated        types.String `tfsdk:"last_updated"`
	Name               types.String `tfsdk:"name"`
	ReservedCount      types.Int64  `tfsdk:"reserved_count"`
	ReservedStatus     types.String `tfsdk:"reserved_status"`
	ServerClass        types.String `tfsdk:"server_class"`
	Taints             types.List   `tfsdk:"taints"`
}

var _ basetypes.ObjectTypable = TaintsType{}

type TaintsType struct {
	basetypes.ObjectType
}

func (t TaintsType) Equal(o attr.Type) bool {
	other, ok := o.(TaintsType)

	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t TaintsType) String() string {
	return "TaintsType"
}

func (t TaintsType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributes := in.Attributes()

	effectAttribute, ok := attributes["effect"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`effect is missing from object`)

		return nil, diags
	}

	effectVal, ok := effectAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`effect expected to be basetypes.StringValue, was: %T`, effectAttribute))
	}

	keyAttribute, ok := attributes["key"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`key is missing from object`)

		return nil, diags
	}

	keyVal, ok := keyAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`key expected to be basetypes.StringValue, was: %T`, keyAttribute))
	}

	valueAttribute, ok := attributes["value"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`value is missing from object`)

		return nil, diags
	}

	valueVal, ok := valueAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`value expected to be basetypes.StringValue, was: %T`, valueAttribute))
	}

	if diags.HasError() {
		return nil, diags
	}

	return TaintsValue{
		Effect: effectVal,
		Key:    keyVal,
		Value:  valueVal,
		state:  attr.ValueStateKnown,
	}, diags
}

func NewTaintsValueNull() TaintsValue {
	return TaintsValue{
		state: attr.ValueStateNull,
	}
}

func NewTaintsValueUnknown() TaintsValue {
	return TaintsValue{
		state: attr.ValueStateUnknown,
	}
}

func NewTaintsValue(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) (TaintsValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/521
	ctx := context.Background()

	for name, attributeType := range attributeTypes {
		attribute, ok := attributes[name]

		if !ok {
			diags.AddError(
				"Missing TaintsValue Attribute Value",
				"While creating a TaintsValue value, a missing attribute value was detected. "+
					"A TaintsValue must contain values for all attributes, even if null or unknown. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("TaintsValue Attribute Name (%s) Expected Type: %s", name, attributeType.String()),
			)

			continue
		}

		if !attributeType.Equal(attribute.Type(ctx)) {
			diags.AddError(
				"Invalid TaintsValue Attribute Type",
				"While creating a TaintsValue value, an invalid attribute value was detected. "+
					"A TaintsValue must use a matching attribute type for the value. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("TaintsValue Attribute Name (%s) Expected Type: %s\n", name, attributeType.String())+
					fmt.Sprintf("TaintsValue Attribute Name (%s) Given Type: %s", name, attribute.Type(ctx)),
			)
		}
	}

	for name := range attributes {
		_, ok := attributeTypes[name]

		if !ok {
			diags.AddError(
				"Extra TaintsValue Attribute Value",
				"While creating a TaintsValue value, an extra attribute value was detected. "+
					"A TaintsValue must not contain values beyond the expected attribute types. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Extra TaintsValue Attribute Name: %s", name),
			)
		}
	}

	if diags.HasError() {
		return NewTaintsValueUnknown(), diags
	}

	effectAttribute, ok := attributes["effect"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`effect is missing from object`)

		return NewTaintsValueUnknown(), diags
	}

	effectVal, ok := effectAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`effect expected to be basetypes.StringValue, was: %T`, effectAttribute))
	}

	keyAttribute, ok := attributes["key"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`key is missing from object`)

		return NewTaintsValueUnknown(), diags
	}

	keyVal, ok := keyAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`key expected to be basetypes.StringValue, was: %T`, keyAttribute))
	}

	valueAttribute, ok := attributes["value"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`value is missing from object`)

		return NewTaintsValueUnknown(), diags
	}

	valueVal, ok := valueAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`value expected to be basetypes.StringValue, was: %T`, valueAttribute))
	}

	if diags.HasError() {
		return NewTaintsValueUnknown(), diags
	}

	return TaintsValue{
		Effect: effectVal,
		Key:    keyVal,
		Value:  valueVal,
		state:  attr.ValueStateKnown,
	}, diags
}

func NewTaintsValueMust(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) TaintsValue {
	object, diags := NewTaintsValue(attributeTypes, attributes)

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

		panic("NewTaintsValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}

	return object
}

func (t TaintsType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.Type() == nil {
		return NewTaintsValueNull(), nil
	}

	if !in.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), in.Type())
	}

	if !in.IsKnown() {
		return NewTaintsValueUnknown(), nil
	}

	if in.IsNull() {
		return NewTaintsValueNull(), nil
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

	return NewTaintsValueMust(TaintsValue{}.AttributeTypes(ctx), attributes), nil
}

func (t TaintsType) ValueType(ctx context.Context) attr.Value {
	return TaintsValue{}
}

var _ basetypes.ObjectValuable = TaintsValue{}

type TaintsValue struct {
	Effect basetypes.StringValue `tfsdk:"effect"`
	Key    basetypes.StringValue `tfsdk:"key"`
	Value  basetypes.StringValue `tfsdk:"value"`
	state  attr.ValueState
}

func (v TaintsValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	attrTypes := make(map[string]tftypes.Type, 3)

	var val tftypes.Value
	var err error

	attrTypes["effect"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["key"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["value"] = basetypes.StringType{}.TerraformType(ctx)

	objectType := tftypes.Object{AttributeTypes: attrTypes}

	switch v.state {
	case attr.ValueStateKnown:
		vals := make(map[string]tftypes.Value, 3)

		val, err = v.Effect.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["effect"] = val

		val, err = v.Key.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["key"] = val

		val, err = v.Value.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["value"] = val

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

func (v TaintsValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v TaintsValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v TaintsValue) String() string {
	return "TaintsValue"
}

func (v TaintsValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	objVal, diags := types.ObjectValue(
		map[string]attr.Type{
			"effect": basetypes.StringType{},
			"key":    basetypes.StringType{},
			"value":  basetypes.StringType{},
		},
		map[string]attr.Value{
			"effect": v.Effect,
			"key":    v.Key,
			"value":  v.Value,
		})

	return objVal, diags
}

func (v TaintsValue) Equal(o attr.Value) bool {
	other, ok := o.(TaintsValue)

	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state != attr.ValueStateKnown {
		return true
	}

	if !v.Effect.Equal(other.Effect) {
		return false
	}

	if !v.Key.Equal(other.Key) {
		return false
	}

	if !v.Value.Equal(other.Value) {
		return false
	}

	return true
}

func (v TaintsValue) Type(ctx context.Context) attr.Type {
	return TaintsType{
		basetypes.ObjectType{
			AttrTypes: v.AttributeTypes(ctx),
		},
	}
}

func (v TaintsValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"effect": basetypes.StringType{},
		"key":    basetypes.StringType{},
		"value":  basetypes.StringType{},
	}
}
