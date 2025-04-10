package mixin

import (
	"github.com/ncobase/ncore/pkg/types"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// FormBaseMixin defines common form-related fields
type FormBaseMixin struct {
	mixin.Schema
}

func (FormBaseMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("form_code").Comment("Form type code"),
		field.String("form_version").Optional().Comment("Form version number"),
		field.JSON("form_config", types.JSON{}).Optional().Comment("Form configuration"),
		field.JSON("form_permissions", types.JSON{}).Optional().Comment("Form permission settings"),
		field.JSON("field_permissions", types.JSON{}).Optional().Comment("Field level permissions"),
	}
}

// NodeBaseMixin defines common node-related fields
type NodeBaseMixin struct {
	mixin.Schema
}

func (NodeBaseMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("node_key").Unique().Comment("Unique identifier for the node"),
		field.String("node_type").Comment("Node type"),
		field.JSON("node_config", types.JSON{}).Optional().Comment("Node configuration"),
		field.JSON("node_rules", types.JSON{}).Optional().Comment("Node rules"),
		field.JSON("node_events", types.JSON{}).Optional().Comment("Node events"),
	}
}

// ProcessRefMixin defines common process reference fields
type ProcessRefMixin struct {
	mixin.Schema
}

func (ProcessRefMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("process_id").Comment("Process instance ID"),
		field.String("template_id").Comment("Process template ID"),
		field.String("business_key").Comment("Business document ID"),
	}
}

// TaskAssigneeMixin defines common task assignee fields
type TaskAssigneeMixin struct {
	mixin.Schema
}

func (TaskAssigneeMixin) Fields() []ent.Field {
	return []ent.Field{
		field.JSON("assignees", types.StringArray{}).Comment("Task assignees"),
		field.JSON("candidates", types.StringArray{}).Comment("Candidate assignees"),
		field.String("delegated_from").Optional().Comment("Delegated from user"),
		field.String("delegated_reason").Optional().Comment("Delegation reason"),
		field.Bool("is_delegated").Default(false).Comment("Whether task is delegated"),
		field.Bool("is_transferred").Default(false).Comment("Whether task is transferred"),
	}
}

// WorkflowControlMixin defines common workflow control fields
type WorkflowControlMixin struct {
	mixin.Schema
}

func (WorkflowControlMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Bool("allow_cancel").Default(true).Comment("Allow cancellation"),
		field.Bool("allow_urge").Default(true).Comment("Allow urging"),
		field.Bool("allow_delegate").Default(true).Comment("Allow delegation"),
		field.Bool("allow_transfer").Default(true).Comment("Allow transfer"),
		field.Bool("is_draft_enabled").Default(true).Comment("Whether draft is enabled"),
		field.Bool("is_auto_start").Default(false).Comment("Whether auto start is enabled"),
		field.Bool("strict_mode").Default(false).Comment("Enable strict mode"),
	}
}

// TimeTrackingMixin defines common time tracking fields
type TimeTrackingMixin struct {
	mixin.Schema
}

func (TimeTrackingMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("start_time").DefaultFunc(func() int64 {
			return time.Now().UnixMilli()
		}).Comment("Start time"),
		field.Int64("end_time").Optional().Nillable().Comment("End time"),
		field.Int64("due_time").Optional().Nillable().Comment("Due time"),
		field.Int("duration").Optional().Comment("Duration in seconds"),
		field.Int("priority").Default(0).Comment("Priority level"),
		field.Bool("is_timeout").Default(false).Comment("Whether timed out"),
		field.Int("reminder_count").Default(0).Comment("Number of reminders sent"),
	}
}

// DataTrackingMixin defines common data tracking fields
type DataTrackingMixin struct {
	mixin.Schema
}

func (DataTrackingMixin) Fields() []ent.Field {
	return []ent.Field{
		field.JSON("origin_data", types.JSON{}).Comment("Original form data"),
		field.JSON("current_data", types.JSON{}).Comment("Current form data"),
		field.JSON("change_logs", []types.JSON{}).Optional().Comment("Data change history"),
		field.Int64("last_modified").Optional().Comment("Last modification time"),
		field.String("last_modifier").Optional().Comment("Last modifier"),
		field.JSON("operation_logs", []types.JSON{}).Optional().Comment("Operation logs"),
	}
}

// BusinessFlowMixin defines common business flow status fields
type BusinessFlowMixin struct {
	mixin.Schema
}

func (BusinessFlowMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("flow_status").Optional().Comment("Flow status"),
		field.JSON("flow_variables", types.JSON{}).Optional().Comment("Flow variables"),
		field.Bool("is_draft").Default(false).Comment("Whether is draft"),
		field.Bool("is_terminated").Default(false).Comment("Whether is terminated"),
		field.Bool("is_suspended").Default(false).Comment("Whether is suspended"),
		field.String("suspend_reason").Optional().Comment("Suspension reason"),
	}
}

// BusinessTagMixin defines common business tagging fields
type BusinessTagMixin struct {
	mixin.Schema
}

func (BusinessTagMixin) Fields() []ent.Field {
	return []ent.Field{
		field.JSON("business_tags", types.StringArray{}).Optional().Comment("Business tags"),
		field.String("module_code").Comment("Module code"),
		field.String("category").Optional().Comment("Category"),
	}
}

// PermissionMixin defines common permission fields
type PermissionMixin struct {
	mixin.Schema
}

func (PermissionMixin) Fields() []ent.Field {
	return []ent.Field{
		field.JSON("viewers", types.StringArray{}).Optional().Comment("Users with view permission"),
		field.JSON("editors", types.StringArray{}).Optional().Comment("Users with edit permission"),
		field.JSON("permission_configs", types.JSON{}).Optional().Comment("Permission configurations"),
		field.JSON("role_configs", types.JSON{}).Optional().Comment("Role configurations"),
		field.JSON("visible_range", types.JSON{}).Optional().Comment("Visibility range"),
	}
}
