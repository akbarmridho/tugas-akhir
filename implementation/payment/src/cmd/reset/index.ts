import { env } from "../../infrastructure/env.js";
import { newRedisCluster } from "../../infrastructure/redis.js";
import { createLogger } from "../../utils/logger.js";

(async function main() {
	const config = env;
	const redis = newRedisCluster(config);

	await redis.flushall();
})().catch((e) => {
	console.error("Main loop error", e);
	process.exit(1);
});
