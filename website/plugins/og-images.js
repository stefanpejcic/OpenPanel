"use strict";Object.defineProperty(exports, "__esModule", {value: true}); function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; } function _nullishCoalesce(lhs, rhsFn) { if (lhs != null) { return lhs; } else { return rhsFn(); } } function _optionalChain(ops) { let lastAccessLHS = undefined; let value = ops[0]; let i = 1; while (i < ops.length) { const op = ops[i]; const fn = ops[i + 1]; i += 2; if ((op === 'optionalAccess' || op === 'optionalCall') && value == null) { return undefined; } if (op === 'access' || op === 'optionalAccess') { lastAccessLHS = value; value = fn(value); } else if (op === 'call' || op === 'optionalCall') { value = fn((...args) => value.call(lastAccessLHS, ...args)); lastAccessLHS = undefined; } } return value; }var _fsextra = require('fs-extra'); var _fsextra2 = _interopRequireDefault(_fsextra);
var _path = require('path'); var _path2 = _interopRequireDefault(_path);
var _crypto = require('crypto'); var _crypto2 = _interopRequireDefault(_crypto);

var _satori = require('satori'); var _satori2 = _interopRequireDefault(_satori);
var _resvgjs = require('@resvg/resvg-js');
















const OG_WIDTH = 1200;
const OG_HEIGHT = 630;
const BACKGROUND_COLOR = "#0B0F17";
const ACCENT_COLOR = "#1890FF";
const MUTED_COLOR = "#9AA4B2";

// static/icons/openpanel_logo.svg, viewBox "0 0 213 215"
const LOGO_PATH =
    "M990 2071 c-39 -13 -141 -66 -248 -129 -53 -32 -176 -103 -272 -158 -206 -117 -276 -177 -306 -264 -17 -50 -19 -88 -19 -460 0 -476 0 -474 94 -568 55 -56 124 -98 604 -369 169 -95 256 -104 384 -37 104 54 532 303 608 353 76 50 126 113 147 184 8 30 12 160 12 447 0 395 -1 406 -22 461 -34 85 -98 138 -317 264 -104 59 -237 136 -295 170 -153 90 -194 107 -275 111 -38 2 -81 0 -95 -5z m205 -561 c66 -38 166 -95 223 -127 l102 -58 0 -262 c0 -262 0 -263 -22 -276 -13 -8 -52 -31 -88 -51 -36 -21 -126 -72 -200 -115 l-135 -78 -3 261 -3 261 -166 95 c-91 52 -190 109 -219 125 -30 17 -52 34 -51 39 3 9 424 256 437 255 3 0 59 -31 125 -69z";

// Keep in sync with permalinkToOgImagePath() in src/theme/DocItem/Metadata/index.js
function permalinkToOgImagePath(permalink) {
    const clean = permalink.replace(/\/+$/, "") || "/index";
    return `/img/og${clean}.png`;
}

// Strip emoji/symbol glyphs the Inter font can't render (would show as tofu boxes).
const EMOJI_PATTERN =
    /[\u{2190}-\u{2BFF}\u{1F000}-\u{1FFFF}\u{FE0F}\u{200D}]/gu;

function truncate(rawText, maxLength) {
    const text = rawText.replace(EMOJI_PATTERN, "").replace(/\s+/g, " ").trim();
    if (text.length <= maxLength) {
        return text;
    }
    return `${text.slice(0, maxLength).trimEnd()}…`;
}








let fontsPromise = null;

function loadFonts() {
    if (!fontsPromise) {
        fontsPromise = Promise.all([
            _fsextra2.default.readFile(
                require.resolve(
                    "@fontsource/inter/files/inter-latin-400-normal.woff",
                ),
            ),
            _fsextra2.default.readFile(
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
    title,
    description,
) {
    const fonts = await loadFonts();

    const svg = await _satori2.default.call(void 0, 
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

    const resvg = new (0, _resvgjs.Resvg)(svg, { fitTo: { mode: "width", value: OG_WIDTH } });
    return resvg.render().asPng();
}



async function readCache(cacheFile) {
    try {
        return await _fsextra2.default.readJson(cacheFile);
    } catch (e) {
        return {};
    }
}

 function pluginOgImages(context

) {
    return {
        name: "openpanel-plugin-og-images",
        async contentLoaded({ allContent }) {
            const docsContent = allContent[
                "docusaurus-plugin-content-docs"
            ] ;

            const currentVersion = _optionalChain([docsContent, 'optionalAccess', _ => _.default, 'access', _2 => _2.loadedVersions, 'access', _3 => _3.find, 'call', _4 => _4(
                (version) => version.versionName === "current",
            )]);

            if (!currentVersion) {
                return;
            }

            const cacheFile = _path2.default.join(
                context.siteDir,
                "node_modules",
                ".cache",
                "og-images.json",
            );

            const cache = await readCache(cacheFile);
            const nextCache = {};

            for (const doc of currentVersion.docs) {
                if (_optionalChain([doc, 'access', _5 => _5.frontMatter, 'optionalAccess', _6 => _6.image])) {
                    // Respect a manually set image, don't generate one.
                    continue;
                }

                const title = truncate(doc.title || "OpenPanel Docs", 80);
                const description = truncate(_nullishCoalesce(doc.description, () => ( "")), 160);
                const imagePath = permalinkToOgImagePath(doc.permalink);
                const outputFile = _path2.default.join(
                    context.siteDir,
                    "static",
                    imagePath,
                );
                const hash = _crypto2.default
                    .createHash("sha1")
                    .update(`${title} ${description}`)
                    .digest("hex");

                nextCache[imagePath] = hash;

                if (
                    cache[imagePath] === hash &&
                    (await _fsextra2.default.pathExists(outputFile))
                ) {
                    continue;
                }

                const png = await renderOgImage(title, description);
                await _fsextra2.default.ensureDir(_path2.default.dirname(outputFile));
                await _fsextra2.default.writeFile(outputFile, png);
            }

            await _fsextra2.default.ensureDir(_path2.default.dirname(cacheFile));
            await _fsextra2.default.writeJson(cacheFile, nextCache);
        },
    };
} exports.default = pluginOgImages;
