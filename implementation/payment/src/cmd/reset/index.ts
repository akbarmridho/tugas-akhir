import { env } from "../../infrastructure/env.js";
import { newRedisCluster } from "../../infrastructure/redis.js";
import { createLogger } from "../../utils/logger.js";

(async function main() {
	const config = env;
	const redis = newRedisCluster(config);
	const logger = createLogger(config);

	logger.info("begin flushing");
	await redis.flushall();
	logger.info("finished flushing");
	process.exit(0);
})().catch((e) => {
	console.error("Main loop error", e);
	process.exit(1);
});
