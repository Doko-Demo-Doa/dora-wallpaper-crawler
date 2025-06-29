import * as cheerio from "cheerio";

const ROOT_LINK = "https://dora-world.com/wallpaper";

export async function getDesktopLinks() {
	try {
		const resp = await fetch(ROOT_LINK);
		const text = await resp.text();

		const $ = cheerio.load(text);
		const desktopSection = $("section.gn_box.wallpaper.nosp > .wallpaper_box");

		const arr: Array<string> = [];

		desktopSection.children().map((_, el) => {
			const link = $(el).children("a").attr("href");
			arr.push(link?.toString() || "");

			return link;
		});

		return arr;
	} catch (error) {
		console.info("Error: ", error);
		return [];
	}
}

export async function getMobileLinks() {
	try {
		const resp = await fetch(ROOT_LINK);
		const text = await resp.text();

		const $ = cheerio.load(text);
		const desktopSection = $("section.gn_box.wallpaper.nopc > .wallpaper_box");

		const arr: Array<string> = [];

		desktopSection.children().map((_, el) => {
			const link = $(el).children("a").attr("href");
			arr.push(link?.toString() || "");

			return link;
		});

		return arr;
	} catch (error) {
		console.info("Error: ", error);
		return [];
	}
}
