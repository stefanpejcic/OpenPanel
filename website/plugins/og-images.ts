import fs from "fs-extra";
import path from "path";
import crypto from "crypto";
import type { Plugin } from "@docusaurus/types";
import satori from "satori";
import { Resvg } from "@resvg/resvg-js";

type DocusaurusDoc = {
    permalink: string;
    title: string;
    description: string;
    frontMatter: {
        image?: string;
    };
};

type ContentPluginType = {
    default: {
        loadedVersions: Array<{ versionName: string; docs: DocusaurusDoc[] }>;
    };
};

const OG_WIDTH = 1200;
const OG_HEIGHT = 630;
const BACKGROUND_COLOR = "#0B0F17";
const ACCENT_COLOR = "#1890FF";
const MUTED_COLOR = "#9AA4B2";

// static/icons/openpanel_logo.svg, viewBox "0 0 213 215"
const LOGO_PATH =
    "M990 2071 c-39 -13 -141 -66 -248 -129 -53 -32 -176 -103 -272 -158 -206 -117 -276 -177 -306 -264 -17 -50 -19 -88 -19 -460 0 -476 0 -474 94 -568 55 -56 124 -98 604 -369 169 -95 256 -104 384 -37 104 54 532 303 608 353 76 50 126 113 147 184 8 30 12 160 12 447 0 395 -1 406 -22 461 -34 85 -98 138 -317 264 -104 59 -237 136 -295 170 -153 90 -194 107 -275 111 -38 2 -81 0 -95 -5z m205 -561 c66 -38 166 -95 223 -127 l102 -58 0 -262 c0 -262 0 -263 -22 -276 -13 -8 -52 -31 -88 -51 -36 -21 -126 -72 -200 -115 l-135 -78 -3 261 -3 261 -166 95 c-91 52 -190 109 -219 125 -30 17 -52 34 -51 39 3 9 424 256 437 255 3 0 59 -31 125 -69z";

// Keep in sync with permalinkToOgImagePath() in src/theme/DocItem/Metadata/index.js
function permalinkToOgImagePath(permalink: string): string {
    const clean = permalink.replace(/\/+$/, "") || "/index";
    return `/img/og${clean}.png`;
}

// Strip emoji/symbol glyphs the Inter font can't render (would show as tofu boxes).
const EMOJI_PATTERN =
    /[\u{2190}-\u{2BFF}\u{1F000}-\u{1FFFF}\u{FE0F}\u{200D}]/gu;

function truncate(rawText: string, maxLength: number): string {
    const text = rawText.replace(EMOJI_PATTERN, "").replace(/\s+/g, " ").trim();
    if (text.length <= maxLength) {
        return text;
    }
    return `${text.slice(0, maxLength).trimEnd()}…`;
}

type FontConfig = {
    name: string;
    data: Buffer;
    weight: 400 | 700;
    style: "normal";
};

let fontsPromise: Promise<FontConfig[]> | null = null;

function loadFonts(): Promise<FontConfig[]> {
    if (!fontsPromise) {
        fontsPromise = Promise.all([
            fs.readFile(
                require.resolve(
                    "@fontsource/inter/files/inter-latin-400-normal.woff",
                ),
            ),
            fs.readFile(
                require.resolve(
                    "@fontsource/inter/files/inter-latin-700-normal.woff",
                ),
            ),
        ]).then(([regular, bold]) => [
            { name: "Inter", data: regular, weight: 400, style: "normal" },
            { name: "Inter", data: bold, weight: 700, style: "normal" },
        ]);
    }
    return fontsPromise;
}

async function renderOgImage(
    title: string,
    description: string,
): Promise<Buffer> {
    const fonts = await loadFonts();

    const svg = await satori(
        {
            type: "div",
            props: {
                style: {
                    height: "100%",
                    width: "100%",
                    display: "flex",
                    flexDirection: "column",
                    justifyContent: "space-between",
                    backgroundColor: BACKGROUND_COLOR,
                    padding: "72px",
                    fontFamily: "Inter",
                },
                children: [
                    {
                        type: "div",
                        props: {
                            style: {
                                display: "flex",
                                alignItems: "center",
                                gap: "16px",
                            },
                            children: [
                                {
                                    type: "svg",
                                    props: {
                                        width: 44,
                                        height: 44,
                                        viewBox: "0 0 213 215",
                                        children: {
                                            type: "path",
                                            props: {
                                                d: LOGO_PATH,
                                                transform:
                                                    "translate(0,215) scale(0.1,-0.1)",
                                                fill: ACCENT_COLOR,
                                            },
                                        },
                                    },
                                },
                                {
                                    type: "span",
                                    props: {
                                        style: {
                                            fontSize: 28,
                                            fontWeight: 700,
                                            color: "#FFFFFF",
                                        },
                                        children: "OpenPanel",
                                    },
                                },
                            ],
                        },
                    },
                    {
                        type: "div",
                        props: {
                            style: { display: "flex", flexDirection: "column" },
                            children: [
                                {
                                    type: "div",
                                    props: {
                                        style: {
                                            display: "-webkit-box",
                                            WebkitBoxOrient: "vertical",
                                            WebkitLineClamp: 2,
                                            textOverflow: "ellipsis",
                                            overflow: "hidden",
                                            fontSize: 58,
                                            fontWeight: 700,
                                            lineHeight: 1.15,
                                            color: "#FFFFFF",
                                            maxWidth: 1000,
                                        },
                                        children: title,
                                    },
                                },
                                ...(description
                                    ? [
                                          {
                                              type: "div",
                                              props: {
                                                  style: {
                                                      display: "-webkit-box",
                                                      WebkitBoxOrient:
                                                          "vertical",
                                                      WebkitLineClamp: 3,
                                                      textOverflow: "ellipsis",
                                                      overflow: "hidden",
                                                      marginTop: 20,
                                                      fontSize: 28,
                                                      lineHeight: 1.4,
                                                      color: MUTED_COLOR,
                                                      maxWidth: 960,
                                                  },
                                                  children: description,
                                              },
                                          },
                                      ]
                                    : []),
                            ],
                        },
                    },
                    {
                        type: "div",
                        props: {
                            style: {
                                display: "flex",
                                fontSize: 22,
                                color: ACCENT_COLOR,
                            },
                            children: "openpanel.com",
                        },
                    },
                ],
            },
        },
        { width: OG_WIDTH, height: OG_HEIGHT, fonts },
    );

    const resvg = new Resvg(svg, { fitTo: { mode: "width", value: OG_WIDTH } });
    return resvg.render().asPng();
}

type CacheMap = Record<string, string>;

async function readCache(cacheFile: string): Promise<CacheMap> {
    try {
        return await fs.readJson(cacheFile);
    } catch {
        return {};
    }
}

export default function pluginOgImages(context: {
    siteDir: string;
}): Plugin<void> {
    return {
        name: "openpanel-plugin-og-images",
        async contentLoaded({ allContent }) {
            const docsContent = allContent[
                "docusaurus-plugin-content-docs"
            ] as ContentPluginType | undefined;

            const currentVersion = docsContent?.default.loadedVersions.find(
                (version) => version.versionName === "current",
            );

            if (!currentVersion) {
                return;
            }

            const cacheFile = path.join(
                context.siteDir,
                "node_modules",
                ".cache",
                "og-images.json",
            );

            const cache = await readCache(cacheFile);
            const nextCache: CacheMap = {};

            for (const doc of currentVersion.docs) {
                if (doc.frontMatter?.image) {
                    // Respect a manually set image, don't generate one.
                    continue;
                }

                const title = truncate(doc.title || "OpenPanel Docs", 80);
                const description = truncate(doc.description ?? "", 160);
                const imagePath = permalinkToOgImagePath(doc.permalink);
                const outputFile = path.join(
                    context.siteDir,
                    "static",
                    imagePath,
                );
                const hash = crypto
                    .createHash("sha1")
                    .update(`${title} ${description}`)
                    .digest("hex");

                nextCache[imagePath] = hash;

                if (
                    cache[imagePath] === hash &&
                    (await fs.pathExists(outputFile))
                ) {
                    continue;
                }

                const png = await renderOgImage(title, description);
                await fs.ensureDir(path.dirname(outputFile));
                await fs.writeFile(outputFile, png);
            }

            await fs.ensureDir(path.dirname(cacheFile));
            await fs.writeJson(cacheFile, nextCache);
        },
    };
}
