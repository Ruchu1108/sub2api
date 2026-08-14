-- 用户默认金额（default_amount）：批量重置余额时恢复到该值；
-- 管理员创建用户时默认取全局 default_balance 设置（见 settings 表）。
-- 幂等：可重复执行而不产生错误。
ALTER TABLE users ADD COLUMN IF NOT EXISTS default_amount decimal(20,8) NOT NULL DEFAULT 0;
