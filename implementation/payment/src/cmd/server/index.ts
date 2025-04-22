import { serve } from "@hono/node-server";
import { createSecureServer } from "node:http2";

import { env } from "../../infrastructure/env.js";
import { readFileSync } from "node:fs";
import { createLogger } from "../../utils/logger.js";
import { newRedisCluster } from "../../infrastructure/redis.js";
import { createWebhookQueue } from "../../infrastructure/queue.js";
import { createApp } from "./app.js";

(async function main() {
	const config = env;
	const logger = createLogger(config);
	const redis = newRedisCluster(config);
	const webhookQueue = createWebhookQueue(redis);

	const app = createApp(logger, redis, webhookQueue);

	const server = serve(
		{
			fetch: app.fetch,
			port: env.PORT,
			createServer: createSecureServer,
			serverOptions: {
				key: readFileSync(config.KEY_PATH),
				cert: readFileSync(config.CERT_PATH),
			},
		},
		(info) => {
			logger.info(`Server is running on ${info.address}:${info.port}`);
		},
	);

	const handleShutdown = (signal: NodeJS.Signals) => {
		logger.info(`Received ${signal}. Shutting down gracefully...`);

		// Close the server
		server.close(() => {
			logger.info("Server closed.");
			webhookQueue.close().finally(() => {
				// Disconnect redis
				redis.disconnect(false);
				process.exit(0);
			});
		});

		// Force close the server after a timeout
		setTimeout(() => {
			logger.error(
				"Could not close connections in time, forcefully shutting down",
			);
			process.exit(1);
		}, 5000); // 5 seconds timeout
	};

	process.on("SIGINT", handleShutdown);
	process.on("SIGTERM", handleShutdown);
})().catch((e) => {
	console.error("Main loop error", e);
	process.exit(1);
});
