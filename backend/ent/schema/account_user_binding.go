package schema

import (
	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// AccountUserBinding 记录账号与用户之间的绑定关系（一对多：一个账号可绑定多个用户）。
//
// 用途：仅作为「重置联动」标记——OpenAI 账号重置限流（消耗 reset credit）成功后，
// 自动把该账号绑定用户的余额重置为各自默认金额。绑定关系不改变账号调度，
// 账号仍然按分组池化共享。
type AccountUserBinding struct {
	ent.Schema
}

// Annotations 指定数据库表名为 "account_user_bindings"。
func (AccountUserBinding) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "account_user_bindings"},
	}
}

// Mixin 自动管理 created_at / updated_at 时间戳。绑定解绑为物理删除，不做软删除。
func (AccountUserBinding) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
	}
}

func (AccountUserBinding) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("account_id"),
		field.Int64("user_id"),
	}
}

func (AccountUserBinding) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("account", Account.Type).
			Ref("user_bindings").
			Field("account_id").
			Unique().
			Required(),
		edge.From("user", User.Type).
			Ref("account_bindings").
			Field("user_id").
			Unique().
			Required(),
	}
}

func (AccountUserBinding) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("account_id", "user_id").Unique(),
		index.Fields("user_id"),
	}
}
