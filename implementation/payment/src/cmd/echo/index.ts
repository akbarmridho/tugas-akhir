import { HTTPException } from "hono/http-exception";
import { logger as honoLogger } from "hono/logger";
import { serve } from "@hono/node-server";
import { timeout } from "hono/timeout";
import { Hono } from "hono";
import { createLogger } from "../../utils/logger.js";
import { env } from "../../infrastructure/env.js";

export const app = new Hono();

const logger = createLogger(env);

app.use("*", timeout(30000));

app.use(
	honoLogger((message, meta) => {
		logger.info(message, meta);
	}),
);

app.notFound((c) => c.json({ message: "Not Found", ok: false }, 404));

app.onError((err, c) => {
	const status = err instanceof HTTPException ? err.status : 500;
	logger.error(`${c.req.method} ${c.req.path}`, {
		error: err,
		status,
		message: err.message,
	});
	c.status(status);
	return c.json({ message: err.message });
});

app.all("*", async (c) => {
	try {
		const contentType = c.req.header("content-type") || "text/plain";

		const body = await c.req.text();

		logger.info("Echo request", {
			headers: c.req.header(),
			body: body,
		});

		c.header("Content-Type", contentType);

		return c.body(body);
	} catch (error) {
		return c.text(`Error processing request: ${(error as Error).message}`, 500);
	}
});

(async function main() {
	const server = serve(
		{
			fetch: app.fetch,
			port: 3005,
		},
		(info) => {
			logger.info(`Echo server is running on ${info.address}:${info.port}`);
		},
	);

	const handleShutdown = (signal: NodeJS.Signals) => {
		logger.info(`Received ${signal}. Shutting down gracefully...`);

		// Close the server
		server.close(() => {
			logger.info("Server closed.");
			process.exit(0);
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
