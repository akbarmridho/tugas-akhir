declare module "https://jslib.k6.io/uuid/1.0.0/index.js" {
	export const v4: () => string;
}

declare module "https://jslib.k6.io/k6-utils/1.2.0/index.js" {
	export const randomIntBetween: (a: number, b: number) => number;
}
