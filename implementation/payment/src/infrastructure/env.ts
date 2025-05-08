import { createEnv } from "@t3-oss/env-core";
import { z } from "zod";
import "dotenv/config";

export const env = createEnv({
	server: {
		REDIS_HOSTS: z
			.string()
			.transform((value) => value.split(","))
			.pipe(z.string().array().min(1)),
		REDIS_HOST_MAPS: z
			.string()
			.default("")
			.transform((value) => value.split(",").filter(e => !!e))
			.pipe(z.string().array()),
		REDIS_PASSWORD: z.string().optional(),
		PORT: z.coerce.number().int().min(50).default(3000),
		NODE_ENV: z
			.enum(["development", "production", "test"])
			.default("development"),
		WEBHOOK_URL: z.string().url(),
		WEBHOOK_SECRET: z.string(),
		KEY_PATH: z.string(),
		CERT_PATH: z.string(),
	},

	/**
	 * The prefix that client-side variables must have. This is enforced both at
	 * a type-level and at runtime.
	 */
	clientPrefix: "PUBLIC_",

	client: {},

	/**
	 * What object holds the environment variables at runtime. This is usually
	 * `process.env` or `import.meta.env`.
	 */
	runtimeEnv: process.env,

	/**
	 * By default, this library will feed the environment variables directly to
	 * the Zod validator.
	 *
	 * This means that if you have an empty string for a value that is supposed
	 * to be a number (e.g. `PORT=` in a ".env" file), Zod will incorrectly flag
	 * it as a type mismatch violation. Additionally, if you have an empty string
	 * for a value that is supposed to be a string with a default value (e.g.
	 * `DOMAIN=` in an ".env" file), the default value will never be applied.
	 *
	 * In order to solve these issues, we recommend that all new projects
	 * explicitly specify this option as true.
	 */
	emptyStringAsUndefined: true,
});

export type Env = typeof env;
