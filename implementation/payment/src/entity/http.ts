import { z } from "zod";

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
