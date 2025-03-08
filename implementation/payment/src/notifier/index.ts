import { Worker } from "bullmq";
import { logger } from "../common/logger.js";
import { redis } from "../common/redis.js";
import { env } from "../common/env.js";
import { InvoiceSchema } from "../common/schema.js";
import fetch from "node-fetch";
import http2Wrapper from "http2-wrapper";
import type { Agent as HttpAgent } from "node:http";

const http2Agent = new http2Wrapper.Agent() as unknown as HttpAgent;

(async function main() {
	await redis.connect();

	// Create the worker to process webhook jobs
	const worker = new Worker<{ id: string }, void, "webhook">(
		"{webhook}",
		async (job) => {
			const { id } = job.data;

			const data = await redis.get(`invoices:${id}`);

			if (data === null) {
				throw new Error("Invoice Not Found");
			}

			const rawInvoice = await InvoiceSchema.safeParseAsync(data);

			if (!rawInvoice.success) {
				throw new Error("Cannot parse invoice data from redis");
			}

			const invoice = rawInvoice.data;

			try {
				const response = await fetch(env.WEBHOOK_URL, {
					method: "POST",
					headers: {
						"content-type": "application/json",
					},
					body: JSON.stringify(invoice),
					agent: http2Agent,
				});

				if (!response.ok) {
					const statusCode = response.status;

					let errorBody: string;
					try {
						const body = await response.json();

						// biome-ignore lint/suspicious/noExplicitAny: <explanation>
						if (Object.hasOwn(body as any, "message")) {
							// biome-ignore lint/suspicious/noExplicitAny: <explanation>
							errorBody = (body as any).message;
						} else {
							errorBody = await response.text();
						}
					} catch {
						errorBody = "Could not read error response body";
					}

					// For other status codes, throw a generic error that will trigger retry
					throw new Error(`HTTP Error: ${statusCode} ${response.statusText}`);
				}
			} catch (error) {
				// Log all errors (both network errors and HTTP errors)
				logger.error(`Webhook error: ${(error as Error).message}`, {
					error: error,
					id: id,
				});

				// Re-throw to trigger retry
				throw error;
			}
		},
		{
			connection: redis,
			concurrency: 500, // Process 10 webhooks simultaneously
		},
	);

	// Graceful shutdown
	async function shutdown() {
		await worker.close();
		process.exit(0);
	}

	process.on("SIGTERM", shutdown);
	process.on("SIGINT", shutdown);
})().catch((e) => {
	logger.error("Error", { error: e });
});
