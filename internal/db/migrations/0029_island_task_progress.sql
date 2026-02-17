CREATE TABLE IF NOT EXISTS island_task_progresses (
  commander_id bigint PRIMARY KEY REFERENCES commanders(commander_id) ON DELETE CASCADE,
  week_start_unix bigint NOT NULL DEFAULT 0,
  last_refresh_day_unix bigint NOT NULL DEFAULT 0,
  week_daily_task_num bigint NOT NULL DEFAULT 0,
  trace_task_id bigint NOT NULL DEFAULT 0,
  trace_daily_task_id bigint NOT NULL DEFAULT 0,
  active_tasks jsonb NOT NULL DEFAULT '[]'::jsonb,
  finished_task_ids jsonb NOT NULL DEFAULT '[]'::jsonb,
  future_task_windows jsonb NOT NULL DEFAULT '[]'::jsonb,
  random_task_windows jsonb NOT NULL DEFAULT '[]'::jsonb,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);
