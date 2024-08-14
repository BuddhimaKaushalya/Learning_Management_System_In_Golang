CREATE TABLE "users" (
  "user_id" bigserial PRIMARY KEY,
  "user_name" varchar NOT NULL UNIQUE,
  "first_name" varchar NOT NULL,
  "last_name" varchar NOT NULL,
  "email" varchar  NOT NULL,
  "is_email_verified" boolean NOT NULL DEFAULT false, 
  "hashed_password" varchar NOT NULL,
  "password_changed_at" timestamptz NOT NULL DEFAULT (now()),
  "role" varchar NOT NULL DEFAULT 'student',
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "user_status" (
  "status_id" bigserial NOT NULL,
  "user_id" bigint NOT NULL,
  "active" boolean NOT NULL DEFAULT false,
  "pending" boolean NOT NULL DEFAULT true,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "sessions" (
  "session_id" uuid PRIMARY KEY,
  "user_id" bigint NOT NULL,
  "refresh_token" varchar NOT NULL,
  "user_agent" varchar NOT NULL,
  "client_ip" varchar NOT NULL,
  "is_blocked" boolean NOT NULL DEFAULT false,
  "expires_at" timestamptz NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "courses" (
  "course_id" bigserial PRIMARY KEY,
  "user_id" bigint NOT NULL,
  "title" varchar NOT NULL,
  "description" varchar NOT NULL,
  "image" varchar NOT NULL,
  "catagory" varchar NOT NULL,
  "what_will" jsonb NOT NULL DEFAULT '{}'::jsonb,
  "sequential_access" boolean NOT NULL DEFAULT true,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())
);


CREATE TABLE "material" (
  "material_id" bigserial PRIMARY KEY,
  "course_id" bigint NOT NULL,
  "title" varchar NOT NULL,
  "material_file" varchar NOT NULL,
  "order_number" bigint UNIQUE NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "assignment" (
  "assignment_id" bigserial PRIMARY KEY,
  "course_id" bigint NOT NULL,
  "title" varchar NOT NULL,
  "assignment_file" varchar NOT NULL,
  "due_date" varchar NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "marks" (
  "mark_id" bigserial PRIMARY KEY,
  "course_id" bigint NOT NULL,
  "user_id" bigint NOT NULL,
  "marks" bigint NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "subscriptions" (
  "subscription_id" bigserial PRIMARY KEY,
  "user_id" bigint NOT NULL,
  "course_id" bigint NOT NULL,
  "active" boolean NOT NULL DEFAULT false,
  "pending" boolean NOT NULL DEFAULT false,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "course_progress" (
  "courseprogress_id" bigserial PRIMARY KEY,
  "course_id" bigint NOT NULL,
  "user_id" bigint NOT NULL,
  "progress" bigint NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "verify_emails" (
  "email_id" bigserial PRIMARY KEY,
  "user_id" bigint NOT NULL,
  "email" varchar NOT NULL,
  "secret_code" varchar NOT NULL,
  "is_used" boolean NOT NULL DEFAULT false,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "expired_at" timestamptz NOT NULL DEFAULT (now() + interval '15 minutes')
);


CREATE TABLE "submission" (
  "submission_id" bigserial PRIMARY KEY,
  "assignment_id" bigint NOT NULL,
  "user_id" bigint NOT NULL,
  "submitted" boolean NOT NULL,
  "grade" varchar NOT NULL,
  "resource" varchar NOT NULL,
  "date_of_submission" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL DEFAULT (now())
);


CREATE TABLE "requests" (
  "request_id" bigserial PRIMARY KEY,
  "course_id" bigint NOT NULL,
  "confirm" boolean NOT NULL DEFAULT false,
  "pending" boolean NOT NULL DEFAULT true,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "lesson_completion" (
    "completion_id" bigserial PRIMARY KEY,
    "user_id" bigint NOT NULL,
    "course_id" bigint NOT NULL,
    "material_id" bigint NOT NULL,
    "completed" boolean NOT NULL DEFAULT false,
    "completed_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "profile_pictures" (
  "profile_picture_id" bigserial PRIMARY KEY,
  "user_id" bigint NOT NULL,
  "picture" varchar NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())
);


CREATE TABLE "categories" (
  "category_id" bigserial PRIMARY KEY,
  "category" varchar NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())
);



CREATE INDEX ON "user_status" ("status_id");

CREATE INDEX ON "user_status" ("user_id");

CREATE INDEX ON "user_status" ("status_id", "user_id");

CREATE INDEX ON "courses" ("course_id");

CREATE INDEX ON "courses" ("user_id");

CREATE INDEX ON "courses" ("course_id", "user_id");

CREATE INDEX ON "requests" ("request_id");

CREATE INDEX ON "requests" ("course_id");

CREATE INDEX ON "material" ("material_id");

CREATE INDEX ON "material" ("course_id");

CREATE INDEX ON "material" ("material_id", "course_id");

CREATE INDEX ON "assignment" ("assignment_id");

CREATE INDEX ON "assignment" ("course_id");

CREATE INDEX ON "assignment" ("assignment_id", "course_id");

CREATE INDEX ON "marks" ("mark_id");

CREATE INDEX ON "marks" ("course_id");

CREATE INDEX ON "marks" ("user_id");

CREATE INDEX ON "categories" ("category_id");

CREATE INDEX ON "marks" ("mark_id", "course_id", "user_id");

CREATE INDEX ON "subscriptions" ("subscription_id");

CREATE INDEX ON "subscriptions" ("user_id");

CREATE INDEX ON "subscriptions" ("course_id");

CREATE INDEX ON "subscriptions" ("subscription_id", "user_id", "course_id");

CREATE INDEX ON "course_progress" ("courseprogress_id");

CREATE INDEX ON "course_progress" ("course_id");

CREATE INDEX ON "course_progress" ("user_id");

CREATE INDEX ON "course_progress" ("courseprogress_id", "course_id", "user_id");

CREATE INDEX ON "submission" ("submission_id");

CREATE INDEX ON "submission" ("assignment_id");

CREATE INDEX ON "submission" ("user_id");
CREATE INDEX ON "verify_emails" ("email_id");

CREATE INDEX ON "profile_pictures" ("profile_picture_id");

CREATE INDEX ON "profile_pictures" ("user_id");

CREATE INDEX ON "submission" ("submission_id", "assignment_id", "user_id");


CREATE INDEX ON "lesson_completion" ("completion_id");

CREATE INDEX ON "lesson_completion" ("material_id");

CREATE INDEX ON "lesson_completion" ("course_id", "user_id");



ALTER TABLE "user_status" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("user_id") ON DELETE CASCADE;

ALTER TABLE "sessions" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("user_id") ON DELETE CASCADE;

ALTER TABLE "courses" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("user_id") ON DELETE CASCADE;

ALTER TABLE "material" ADD FOREIGN KEY ("course_id") REFERENCES "courses" ("course_id") ON DELETE CASCADE;

ALTER TABLE "marks" ADD FOREIGN KEY ("course_id") REFERENCES "courses" ("course_id") ON DELETE CASCADE;

ALTER TABLE "marks" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("user_id") ON DELETE CASCADE;

ALTER TABLE "subscriptions" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("user_id") ON DELETE CASCADE;

ALTER TABLE "subscriptions" ADD FOREIGN KEY ("course_id") REFERENCES "courses" ("course_id") ON DELETE CASCADE;

ALTER TABLE "course_progress" ADD FOREIGN KEY ("course_id") REFERENCES "courses" ("course_id") ON DELETE CASCADE;

ALTER TABLE "course_progress" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("user_id") ON DELETE CASCADE;

ALTER TABLE "submission" ADD FOREIGN KEY ("assignment_id") REFERENCES "assignment" ("assignment_id") ON DELETE CASCADE;

ALTER TABLE "submission" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("user_id") ON DELETE CASCADE;

ALTER TABLE "verify_emails" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("user_id") ON DELETE CASCADE;

ALTER TABLE "requests" ADD FOREIGN KEY ("course_id") REFERENCES "courses" ("course_id")  ON DELETE CASCADE;

ALTER TABLE "lesson_completion" ADD FOREIGN KEY ("course_id") REFERENCES "courses" ("course_id")  ON DELETE CASCADE;

ALTER TABLE "lesson_completion" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("user_id")  ON DELETE CASCADE;

ALTER TABLE "lesson_completion" ADD FOREIGN KEY ("material_id") REFERENCES "material" ("material_id")  ON DELETE CASCADE;

ALTER TABLE "profile_pictures" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("user_id")  ON DELETE CASCADE;

ALTER TABLE "assignment" ADD FOREIGN KEY ("course_id") REFERENCES "courses" ("course_id") ON DELETE CASCADE;












