-- 账号↔用户绑定关系：OpenAI 账号重置限流（消耗 reset credit）成功后，
-- 自动把绑定用户的余额重置为各自默认金额（仅作重置联动标记，不改变账号调度）。
-- 幂等：可重复执行而不产生错误。
CREATE TABLE IF NOT EXISTS account_user_bindings (
    id         BIGSERIAL PRIMARY KEY,
    account_id BIGINT NOT NULL REFERENCES accounts (id) ON DELETE CASCADE,
    user_id    BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT account_user_bindings_account_user_unique UNIQUE (account_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_account_user_bindings_user_id
    ON account_user_bindings (user_id);
