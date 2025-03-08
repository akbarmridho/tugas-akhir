import { z } from "@hono/zod-openapi";

export const IdParamsSchema = z.object({
	id: z
		.string()
		.min(3)
		.openapi({
			param: {
				name: "id",
				in: "path",
			},
		}),
});

export const ErrorSchema = z.object({
	message: z.string(),
});

export const InvoiceSchema = z
	.object({
		id: z.string(),
		amount: z.number().gt(0),
		description: z.string().default(""),
		externalId: z.string(),
		createdAt: z.coerce.date(),
		expiredAt: z.coerce.date(),
		paidAt: z.coerce.date().nullable(),
		paidAmount: z.number().nullable(),
		status: z.enum(["pending", "expired", "paid", "failed"]),
	})
	.openapi("Invoice");

export type InvoiceType = z.infer<typeof InvoiceSchema>;

export const CreateInvoiceSchema = z
	.object({
		amount: z.number().gt(0),
		description: z.string().default(""),
		externalId: z.string(),
	})
	.openapi("CreateInvoiceRequest");

export const PayInvoiceSchema = z
	.object({
		mode: z.enum(["success", "failed"]),
	})
	.openapi("PayInvoiceRequest");
