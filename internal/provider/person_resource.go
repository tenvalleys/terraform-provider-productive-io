// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &PersonResource{}
var _ resource.ResourceWithImportState = &PersonResource{}

func NewPersonResource() resource.Resource {
	return &PersonResource{}
}

type PersonResource struct {
	client *Client
}

// PersonResourceModel is the Terraform state model for a productive_person resource.
type PersonResourceModel struct {
	// Identifer
	ID types.String `tfsdk:"id"`

	// Required
	FirstName types.String `tfsdk:"first_name"`
	LastName  types.String `tfsdk:"last_name"`
	Email     types.String `tfsdk:"email"`

	// Optional
	Title                       types.String `tfsdk:"title"`
	Nickname                    types.String `tfsdk:"nickname"`
	TagList                     types.String `tfsdk:"tag_list"`
	RoleID                      types.Int64  `tfsdk:"role_id"`
	CompanyID                   types.Int64  `tfsdk:"company_id"`
	ManagerID                   types.Int64  `tfsdk:"manager_id"`
	SubsidiaryID                types.Int64  `tfsdk:"subsidiary_id"`
	CustomRoleID                types.Int64  `tfsdk:"custom_role_id"`
	TimeTrackingPolicyID        types.Int64  `tfsdk:"time_tracking_policy_id"`
	TimesheetSubmissionDisabled types.Bool   `tfsdk:"timesheet_submission_disabled"`
	Virtual                     types.Bool   `tfsdk:"virtual"`

	// Computed (read-only from API)
	Status        types.Int64  `tfsdk:"status"`
	CreatedAt     types.String `tfsdk:"created_at"`
	InvitedAt     types.String `tfsdk:"invited_at"`
	LastSeenAt    types.String `tfsdk:"last_seen_at"`
	IsUser        types.Bool   `tfsdk:"is_user"`
	AvatarURL     types.String `tfsdk:"avatar_url"`
	TwoFactorAuth types.Bool   `tfsdk:"two_factor_auth"`
}

func (r *PersonResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_person"
}

func (r *PersonResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a person (user) in Productive.io. " +
			"Because Productive.io has no hard-delete endpoint, destroying this resource archives the person.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier of the person.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			// --- Required ---
			"first_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "First name of the person.",
			},
			"last_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Last name of the person.",
			},
			"email": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Email address of the person.",
			},

			// --- Optional ---
			"title": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Job title of the person.",
			},
			"nickname": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Preferred display name.",
			},
			"tag_list": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Comma-separated list of tags.",
			},
			"role_id": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Built-in role ID (e.g. 1 = admin).",
			},
			"company_id": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "ID of the associated company.",
			},
			"manager_id": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "ID of the direct manager (person).",
			},
			"subsidiary_id": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "ID of the subsidiary/workplace.",
			},
			"custom_role_id": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "ID of the custom permission role.",
			},
			"time_tracking_policy_id": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "ID of the time tracking policy.",
			},
			"timesheet_submission_disabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether timesheet submission is disabled for this person.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"virtual": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether this is a virtual/placeholder record.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},

			// --- Computed ---
			"status": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Person status: 1 = active, 2 = deactivated.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Timestamp when the person was created (ISO 8601).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"invited_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Timestamp when the invitation was sent (ISO 8601), or null.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_seen_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Timestamp of the person's last activity (ISO 8601), or null.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"is_user": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether this person has a login account.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"avatar_url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "URL of the person's avatar thumbnail.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"two_factor_auth": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether two-factor authentication is enabled.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *PersonResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *PersonResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PersonResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	person, err := r.client.CreatePerson(ctx, modelToWriteAttrs(data))
	if err != nil {
		resp.Diagnostics.AddError("Error creating person", err.Error())
		return
	}

	personDataToModel(person, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PersonResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PersonResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	person, err := r.client.GetPerson(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading person", err.Error())
		return
	}
	if person == nil {
		// Not found — remove from state.
		resp.State.RemoveResource(ctx)
		return
	}
	// Archived persons are treated as destroyed from Terraform's perspective.
	if person.Attributes.ArchivedAt != nil {
		resp.State.RemoveResource(ctx)
		return
	}

	personDataToModel(person, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PersonResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data PersonResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sent := modelToWriteAttrs(data)
	person, err := r.client.UpdatePerson(ctx, data.ID.ValueString(), sent)
	if err != nil {
		resp.Diagnostics.AddError("Error updating person", err.Error())
		return
	}

	if rejected := detectRejectedFields(sent, person.Attributes); len(rejected) > 0 {
		resp.Diagnostics.AddError(
			"API silently rejected field updates",
			fmt.Sprintf(
				"The Productive.io API accepted the request but did not apply changes to: %s. "+
					"This usually means your API token lacks permission to modify these fields. "+
					"Use a token with full admin access or update these fields in the Productive.io UI.",
				strings.Join(rejected, ", "),
			),
		)
		return
	}

	personDataToModel(person, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PersonResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PersonResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.ArchivePerson(ctx, data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error archiving person", err.Error())
	}
}

func (r *PersonResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Fetch the full person from the API on import so all computed fields are populated.
	person, err := r.client.GetPerson(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing person", err.Error())
		return
	}
	if person == nil {
		resp.Diagnostics.AddError("Person not found", fmt.Sprintf("No person with ID %q exists.", req.ID))
		return
	}

	var data PersonResourceModel
	personDataToModel(person, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// --- helpers ---

// modelToWriteAttrs converts the Terraform state model to the API write payload.
func modelToWriteAttrs(data PersonResourceModel) personWriteAttrs {
	attrs := personWriteAttrs{
		FirstName: data.FirstName.ValueString(),
		LastName:  data.LastName.ValueString(),
		Email:     data.Email.ValueString(),
	}

	if !data.Title.IsNull() && !data.Title.IsUnknown() {
		attrs.Title = data.Title.ValueString()
	}
	if !data.Nickname.IsNull() && !data.Nickname.IsUnknown() {
		attrs.Nickname = data.Nickname.ValueString()
	}
	if !data.TagList.IsNull() && !data.TagList.IsUnknown() {
		attrs.TagList = data.TagList.ValueString()
	}
	if !data.RoleID.IsNull() && !data.RoleID.IsUnknown() {
		v := data.RoleID.ValueInt64()
		attrs.RoleID = &v
	}
	if !data.CompanyID.IsNull() && !data.CompanyID.IsUnknown() {
		v := data.CompanyID.ValueInt64()
		attrs.CompanyID = &v
	}
	if !data.ManagerID.IsNull() && !data.ManagerID.IsUnknown() {
		v := data.ManagerID.ValueInt64()
		attrs.ManagerID = &v
	}
	if !data.SubsidiaryID.IsNull() && !data.SubsidiaryID.IsUnknown() {
		v := data.SubsidiaryID.ValueInt64()
		attrs.SubsidiaryID = &v
	}
	if !data.CustomRoleID.IsNull() && !data.CustomRoleID.IsUnknown() {
		v := data.CustomRoleID.ValueInt64()
		attrs.CustomRoleID = &v
	}
	if !data.TimeTrackingPolicyID.IsNull() && !data.TimeTrackingPolicyID.IsUnknown() {
		v := data.TimeTrackingPolicyID.ValueInt64()
		attrs.TimeTrackingPolicyID = &v
	}
	if !data.TimesheetSubmissionDisabled.IsNull() && !data.TimesheetSubmissionDisabled.IsUnknown() {
		v := data.TimesheetSubmissionDisabled.ValueBool()
		attrs.TimesheetSubmissionDisabled = &v
	}
	if !data.Virtual.IsNull() && !data.Virtual.IsUnknown() {
		v := data.Virtual.ValueBool()
		attrs.Virtual = &v
	}

	return attrs
}

// personDataToModel populates the Terraform state model from an API response.
func personDataToModel(person *PersonData, data *PersonResourceModel) {
	data.ID = types.StringValue(person.ID)
	data.FirstName = types.StringValue(person.Attributes.FirstName)
	data.LastName = types.StringValue(person.Attributes.LastName)
	data.Email = types.StringValue(person.Attributes.Email)

	// Optional strings: preserve null when API returns empty.
	setStringOrNull := func(dst *types.String, val string) {
		if val != "" {
			*dst = types.StringValue(val)
		} else {
			*dst = types.StringNull()
		}
	}
	setStringOrNull(&data.Title, person.Attributes.Title)
	setStringOrNull(&data.Nickname, person.Attributes.Nickname)
	setStringOrNull(&data.TagList, strings.Join(person.Attributes.TagList, ","))

	// Optional ints: null when API returns nil.
	setInt64OrNull := func(dst *types.Int64, val *int64) {
		if val != nil {
			*dst = types.Int64Value(*val)
		} else {
			*dst = types.Int64Null()
		}
	}
	setInt64OrNull(&data.RoleID, person.Attributes.RoleID)
	setInt64OrNull(&data.CompanyID, person.Attributes.CompanyID)
	setInt64OrNull(&data.ManagerID, person.Attributes.ManagerID)
	setInt64OrNull(&data.SubsidiaryID, person.Attributes.SubsidiaryID)
	setInt64OrNull(&data.CustomRoleID, person.Attributes.CustomRoleID)
	setInt64OrNull(&data.TimeTrackingPolicyID, person.Attributes.TimeTrackingPolicyID)

	// Optional+Computed bools: always reflect API truth.
	data.TimesheetSubmissionDisabled = types.BoolValue(person.Attributes.TimesheetSubmissionDisabled)
	data.Virtual = types.BoolValue(person.Attributes.Virtual)

	// Computed fields.
	data.Status = types.Int64Value(person.Attributes.Status)
	data.CreatedAt = types.StringValue(person.Attributes.CreatedAt)
	data.IsUser = types.BoolValue(person.Attributes.IsUser)
	data.TwoFactorAuth = types.BoolValue(person.Attributes.TwoFactorAuth)

	setStringOrNull(&data.AvatarURL, person.Attributes.AvatarURL)

	if person.Attributes.InvitedAt != nil {
		data.InvitedAt = types.StringValue(*person.Attributes.InvitedAt)
	} else {
		data.InvitedAt = types.StringNull()
	}
	if person.Attributes.LastSeenAt != nil {
		data.LastSeenAt = types.StringValue(*person.Attributes.LastSeenAt)
	} else {
		data.LastSeenAt = types.StringNull()
	}
}

// detectRejectedFields compares the values sent in the write request against
// what the API returned. Any field that was sent but came back unchanged is
// reported by name. The Productive.io API silently ignores fields the token
// lacks permission to modify (HTTP 200, no error), so this is the only way
// to surface the problem.
func detectRejectedFields(sent personWriteAttrs, got PersonAttributes) []string {
	var rejected []string

	if sent.FirstName != got.FirstName {
		rejected = append(rejected, "first_name")
	}
	if sent.LastName != got.LastName {
		rejected = append(rejected, "last_name")
	}
	if sent.Email != got.Email {
		rejected = append(rejected, "email")
	}
	if sent.Title != "" && sent.Title != got.Title {
		rejected = append(rejected, "title")
	}
	if sent.Nickname != "" && sent.Nickname != got.Nickname {
		rejected = append(rejected, "nickname")
	}
	if sent.TagList != "" && sent.TagList != strings.Join(got.TagList, ",") {
		rejected = append(rejected, "tag_list")
	}
	if sent.RoleID != nil && (got.RoleID == nil || *sent.RoleID != *got.RoleID) {
		rejected = append(rejected, "role_id")
	}
	if sent.CompanyID != nil && (got.CompanyID == nil || *sent.CompanyID != *got.CompanyID) {
		rejected = append(rejected, "company_id")
	}
	if sent.ManagerID != nil && (got.ManagerID == nil || *sent.ManagerID != *got.ManagerID) {
		rejected = append(rejected, "manager_id")
	}
	if sent.SubsidiaryID != nil && (got.SubsidiaryID == nil || *sent.SubsidiaryID != *got.SubsidiaryID) {
		rejected = append(rejected, "subsidiary_id")
	}
	if sent.CustomRoleID != nil && (got.CustomRoleID == nil || *sent.CustomRoleID != *got.CustomRoleID) {
		rejected = append(rejected, "custom_role_id")
	}
	if sent.TimeTrackingPolicyID != nil && (got.TimeTrackingPolicyID == nil || *sent.TimeTrackingPolicyID != *got.TimeTrackingPolicyID) {
		rejected = append(rejected, "time_tracking_policy_id")
	}
	if sent.TimesheetSubmissionDisabled != nil && *sent.TimesheetSubmissionDisabled != got.TimesheetSubmissionDisabled {
		rejected = append(rejected, "timesheet_submission_disabled")
	}
	if sent.Virtual != nil && *sent.Virtual != got.Virtual {
		rejected = append(rejected, "virtual")
	}

	return rejected
}
