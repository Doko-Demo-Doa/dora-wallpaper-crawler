import { inArray } from "drizzle-orm";
import { difference } from "remeda";
import { UTApi } from "uploadthing/server";
import { db } from "../db";
import { type NewWallpaper, wallpapers } from "../db/schema";
import { getDesktopLinks, getMobileLinks } from "./parser";

export const utapi = new UTApi();

export async function startParsingAndDownload() {
	console.info(
		`Starting to parse and download wallpapers at ${new Date().toUTCString()}`,
	);
	const desktopLinks = await getDesktopLinks();
	const mobileLinks = await getMobileLinks();

	const query = await db
		.select()
		.from(wallpapers)
		.where(inArray(wallpapers.originalUrl, [...desktopLinks, ...mobileLinks]));

	// Only get links in desktopLinks that are not in the database
	const diff = difference(
		[...desktopLinks, ...mobileLinks],
		query.map((q) => q.originalUrl),
	);

	if (diff.length) {
		console.info("New links found: ", diff);
	}

	const uploadedFiles = await utapi.uploadFilesFromUrl(diff, {
		acl: "public-read",
	});

	for (const uFile of uploadedFiles) {
		if (!uFile.error) {
			const newWallpaper: NewWallpaper = {
				createdAt: new Date().toUTCString(),
				fileName: uFile.data?.name || "",
				isMobile: 0,
				originalUrl: uFile.data?.ufsUrl || "",
				url: uFile.data?.ufsUrl || "",
			};
			await db.insert(wallpapers).values(newWallpaper);
		}
	}
}
