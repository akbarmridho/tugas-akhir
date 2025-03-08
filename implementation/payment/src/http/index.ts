import { serve } from "@hono/node-server";
import { createServer } from "node:http2";

import { logger } from "../common/logger.js";
import { env } from "../common/env.js";
import { app } from "./app.js";
import { redis } from "../common/redis.js";

(async function main() {
	await redis.connect();

	const server = serve(
		{
			fetch: app.fetch,
			port: env.PORT,
			createServer,
		},
		(info) => {
			logger.info(`Server is running on http://0.0.0.0:${info.port}`);
		},
	);

	const handleShutdown = (signal: NodeJS.Signals) => {
		logger.info(`Received ${signal}. Shutting down gracefully...`);

		// Close the server
		server.close(() => {
			logger.info("Server closed.");
			// Disconnect redis
			redis.disconnect().finally(() => {
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
	logger.error("Main loop error", {
		error: e,
	});
	process.exit(1);
});
