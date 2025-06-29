import { cron, Patterns } from "@elysiajs/cron";
import { swagger } from "@elysiajs/swagger";
import { inArray } from "drizzle-orm";
import { Elysia } from "elysia";
import * as R from "remeda";
import { startParsingAndDownload } from "./crons/downloader";
import { getDesktopLinks } from "./crons/parser";
import { db } from "./db";
import { wallpapers } from "./db/schema";

const app = new Elysia()
	.use(swagger())
	.use(
		cron({
			name: "heartbeat",
			pattern: Patterns.everyHours(),
			run() {
				startParsingAndDownload();
			},
		}),
	)
	.group("/api", (app) =>
		app
			.get("/", async () => {
				const desktopLinks = await getDesktopLinks();

				// Check for existing links
				const query = await db
					.select()
					.from(wallpapers)
					.where(inArray(wallpapers.originalUrl, desktopLinks));

				const diff = R.difference(
					desktopLinks,
					query.map((q) => q.originalUrl),
				);
				return diff;
			})
			.get("/all", async () => {
				const wallpapersData = await db.select().from(wallpapers);
				return wallpapersData;
			}),
	)
	.listen(process.env.PORT ?? 3000);

console.log(
	`🦊 Elysia is running at ${app.server?.hostname}:${app.server?.port}`,
);
