import { serve } from "@hono/node-server";
import { createSecureServer } from "node:http2";

import { logger } from "../common/logger.js";
import { env } from "../common/env.js";
import { app } from "./app.js";
import { redis } from "../common/redis.js";
import { queue } from "./queue.js";
import { readFileSync } from "node:fs";

(async function main() {
	const server = serve(
		{
			fetch: app.fetch,
			port: env.PORT,
			createServer: createSecureServer,
			serverOptions: {
				key: readFileSync("../cert/key.pem"),
				cert: readFileSync("../cert/cert.pem"),
			},
		},
		(info) => {
			logger.info(`Server is running on http://localhost:${info.port}`);
		},
	);

	const handleShutdown = (signal: NodeJS.Signals) => {
		logger.info(`Received ${signal}. Shutting down gracefully...`);

		// Close the server
		server.close(() => {
			logger.info("Server closed.");
			queue.close().finally(() => {
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
	logger.error("Main loop error", e);
	process.exit(1);
});
