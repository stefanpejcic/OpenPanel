import React, { useEffect, useState } from "react";
import clsx from "clsx";
import Head from "@docusaurus/Head";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";

const LOGO_URL = "/img/svg/openpanel_logo.svg";

const VARIANTS = [
    { id: "black", name: "Black", fill: "#000000", bg: "#FFFFFF", note: "On white and light backgrounds" },
    { id: "white", name: "White", fill: "#FFFFFF", bg: "#14141F", note: "On dark backgrounds and photos" },
    { id: "blue", name: "Blue", fill: "#0080FF", bg: "#FFFFFF", note: "Brand color on light backgrounds" },
    { id: "cyan", name: "Cyan", fill: "#47EBEB", bg: "#14141F", note: "Brand color on dark backgrounds" },
];

const COLORS = [
    { name: "OpenPanel Blue", hex: "#0080FF", rgb: "0, 128, 255", text: "#FFFFFF" },
    { name: "Cyan", hex: "#47EBEB", rgb: "71, 235, 235", text: "#14141F" },
    { name: "Night", hex: "#14141F", rgb: "20, 20, 31", text: "#FFFFFF" },
    { name: "White", hex: "#FFFFFF", rgb: "255, 255, 255", text: "#14141F" },
];

const SCREENSHOTS = [
    { name: "OpenPanel dashboard", light: "/img/openpanel-screenshots/dashboard/dashboard-window.png" },
    { name: "OpenAdmin dashboard", light: "/img/openadmin-screenshots/001_dashboard-window.png" },
    { name: "OpenPanel, Modern theme", light: "/img/openpanel-screenshots/themes/modern.png", noDark: true },
    { name: "OpenPanel, Saturn theme", light: "/img/openpanel-screenshots/themes/saturn.png", noDark: true },
];

const PNG_SIZES = [256, 512, 1024];

const card = "rounded-2xl border border-gray-300 dark:border-gray-700 p-5 sm:p-6 flex flex-col gap-5";
const smallBtn =
    "rounded-full px-3 py-1.5 text-xs font-semibold border border-gray-300 dark:border-gray-700 text-gray-800 dark:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-800 cursor-pointer bg-transparent no-underline";
const link = "text-refine-react-light-link dark:text-refine-react-dark-link hover:underline";

const colored = (svg: string, fill: string) => svg.replace(/fill="currentColor"/g, `fill="${fill}"`);

// the icon next to the OpenPanel name, the text needs Inter installed to look right in the SVG
const lockupSvg = (svg: string, fill: string) => {
    const inner = svg.replace(/^[\s\S]*?<svg[^>]*>/, "").replace(/<\/svg>\s*$/, "");
    return `<svg xmlns="http://www.w3.org/2000/svg" width="1240" height="300" viewBox="0 0 1240 300"><svg x="20" y="40" width="220" height="220" viewBox="0 0 213 215">${colored(
        inner,
        fill,
    )}</svg><text x="280" y="196" font-family="Inter, Arial, sans-serif" font-size="150" font-weight="700" fill="${fill}">OpenPanel</text></svg>`;
};

const save = (blob: Blob, name: string) => {
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = name;
    document.body.appendChild(a);
    a.click();
    a.remove();
    URL.revokeObjectURL(url);
};

const toPng = (svg: string, width: number, height: number, name: string) => {
    const img = new Image();
    img.onload = () => {
        const canvas = document.createElement("canvas");
        canvas.width = width;
        canvas.height = height;
        canvas.getContext("2d")!.drawImage(img, 0, 0, width, height);
        canvas.toBlob((b) => b && save(b, name), "image/png");
    };
    img.src = "data:image/svg+xml;charset=utf-8," + encodeURIComponent(svg);
};

const CopyChip = ({ value }: { value: string }) => {
    const [done, setDone] = useState(false);
    return (
        <button
            type="button"
            onClick={async () => {
                try {
                    await navigator.clipboard.writeText(value);
                    setDone(true);
                    setTimeout(() => setDone(false), 1200);
                } catch {
                    setDone(false);
                }
            }}
            className="font-mono text-xs rounded px-1.5 py-0.5 bg-black/10 hover:bg-black/20 border-0 cursor-pointer text-current"
        >
            {done ? "Copied" : value}
        </button>
    );
};

const Assets: React.FC = () => {
    const [svg, setSvg] = useState("");
    const [custom, setCustom] = useState("#0080FF");

    useEffect(() => {
        fetch(LOGO_URL)
            .then((r) => r.text())
            .then(setSvg)
            .catch(() => setSvg(""));
    }, []);

    const icon = (fill: string, size: number) => (
        <div
            className="[&_svg]:w-full [&_svg]:h-full"
            style={{ width: size, height: size }}
            dangerouslySetInnerHTML={{ __html: svg ? colored(svg, fill) : "" }}
        />
    );

    const downloads = (fill: string, id: string) => (
        <div className="flex flex-wrap gap-2">
            <button type="button" className={smallBtn} onClick={() => save(new Blob([colored(svg, fill)], { type: "image/svg+xml" }), `openpanel-icon-${id}.svg`)}>
                SVG
            </button>
            {PNG_SIZES.map((s) => (
                <button key={s} type="button" className={smallBtn} onClick={() => toPng(colored(svg, fill), s, s, `openpanel-icon-${id}-${s}.png`)}>
                    PNG {s}
                </button>
            ))}
        </div>
    );

    return (
        <CommonLayout
            description="Download official OpenPanel brand assets: logo in SVG and PNG, brand colors, font and product screenshots for press, partners and integrations."
            image="/img/og/assets.png"
        >
            <Head>
                <title>Brand Assets | OpenPanel</title>
                <meta property="og:title" content="Brand Assets | OpenPanel" />
                <meta property="og:image:width" content="1200" />
                <meta property="og:image:height" content="630" />
                <meta property="og:image:type" content="image/png" />
                <html data-page="brand-assets" data-customized="true" />
            </Head>
            <div className="refine-prose">
                <CommonHeader hasSticky={true} />

                <div className="not-prose flex flex-col gap-10 pt-8 lg:pt-16 pb-32 max-w-[928px] landing-lg:max-w-[1232px] w-full mx-auto px-4">
                    <section className="flex flex-col gap-3">
                        <h1 className="text-3xl sm:text-5xl font-bold text-gray-900 dark:text-gray-0 m-0">Brand assets</h1>
                        <p className="text-base sm:text-lg text-gray-600 dark:text-gray-400 m-0 max-w-[720px]">
                            Logos, colors and screenshots for articles, reviews, partner pages and integrations. Download what you
                            need, and please keep the logo as it is.
                        </p>
                    </section>

                    <section className={card}>
                        <div className="flex flex-col gap-1">
                            <h2 className="text-xl font-bold text-gray-900 dark:text-gray-0 m-0">Logo</h2>
                            <p className="text-sm text-gray-600 dark:text-gray-400 m-0">Pick the version that has the most contrast with your background.</p>
                        </div>
                        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
                            {VARIANTS.map((v) => (
                                <div key={v.id} className="rounded-xl border border-gray-200 dark:border-gray-700 overflow-hidden flex flex-col">
                                    <div className="h-40 flex items-center justify-center" style={{ background: v.bg }}>
                                        {icon(v.fill, 96)}
                                    </div>
                                    <div className="p-4 flex flex-col gap-3">
                                        <div className="flex flex-col">
                                            <span className="text-sm font-semibold text-gray-900 dark:text-gray-100">{v.name}</span>
                                            <span className="text-xs text-gray-600 dark:text-gray-400">{v.note}</span>
                                        </div>
                                        {downloads(v.fill, v.id)}
                                    </div>
                                </div>
                            ))}
                        </div>
                    </section>

                    <section className="grid gap-6 lg:grid-cols-2">
                        <div className={card}>
                            <div className="flex flex-col gap-1">
                                <h2 className="text-xl font-bold text-gray-900 dark:text-gray-0 m-0">Logo with name</h2>
                                <p className="text-sm text-gray-600 dark:text-gray-400 m-0">
                                    For headers and partner pages. The name is set in Inter Bold.
                                </p>
                            </div>
                            <div className="grid grid-cols-2 gap-3">
                                {[
                                    { fill: "#000000", bg: "#FFFFFF", id: "black" },
                                    { fill: "#FFFFFF", bg: "#14141F", id: "white" },
                                ].map((l) => (
                                    <div key={l.id} className="rounded-xl border border-gray-200 dark:border-gray-700 overflow-hidden flex flex-col">
                                        <div className="h-24 flex items-center justify-center gap-2 px-3" style={{ background: l.bg }}>
                                            {icon(l.fill, 36)}
                                            <span className="text-2xl font-bold" style={{ color: l.fill }}>
                                                OpenPanel
                                            </span>
                                        </div>
                                        <div className="p-3 flex flex-wrap gap-2">
                                            <button type="button" className={smallBtn} onClick={() => save(new Blob([lockupSvg(svg, l.fill)], { type: "image/svg+xml" }), `openpanel-logo-${l.id}.svg`)}>
                                                SVG
                                            </button>
                                            <button type="button" className={smallBtn} onClick={() => toPng(lockupSvg(svg, l.fill), 1240, 300, `openpanel-logo-${l.id}.png`)}>
                                                PNG
                                            </button>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        </div>

                        <div className={card}>
                            <div className="flex flex-col gap-1">
                                <h2 className="text-xl font-bold text-gray-900 dark:text-gray-0 m-0">Any color</h2>
                                <p className="text-sm text-gray-600 dark:text-gray-400 m-0">Need the icon in your own palette? Pick a color and download.</p>
                            </div>
                            <div className="flex items-center gap-5">
                                <div className="rounded-xl border border-gray-200 dark:border-gray-700 p-4 bg-gray-50 dark:bg-gray-800">{icon(custom, 72)}</div>
                                <label className="flex flex-col gap-1.5 text-sm font-semibold text-gray-700 dark:text-gray-300">
                                    Color
                                    <input type="color" value={custom} onChange={(e) => setCustom(e.target.value)} className="h-9 w-16 cursor-pointer rounded border border-gray-300 dark:border-gray-700 bg-transparent" />
                                </label>
                            </div>
                            {downloads(custom, "custom")}
                        </div>
                    </section>

                    <section className={card}>
                        <div className="flex flex-col gap-1">
                            <h2 className="text-xl font-bold text-gray-900 dark:text-gray-0 m-0">Colors and font</h2>
                            <p className="text-sm text-gray-600 dark:text-gray-400 m-0">Click a value to copy it.</p>
                        </div>
                        <div className="grid gap-3 grid-cols-2 lg:grid-cols-4">
                            {COLORS.map((c) => (
                                <div key={c.hex} className="rounded-xl border border-gray-200 dark:border-gray-700 overflow-hidden">
                                    <div className="h-24 p-3 flex flex-col justify-end gap-1" style={{ background: c.hex, color: c.text }}>
                                        <span className="text-sm font-semibold">{c.name}</span>
                                        <div className="flex flex-wrap gap-1.5">
                                            <CopyChip value={c.hex} />
                                            <CopyChip value={`rgb(${c.rgb})`} />
                                        </div>
                                    </div>
                                </div>
                            ))}
                        </div>
                        <div className="rounded-xl bg-gray-50 dark:bg-gray-800/60 p-4 flex flex-col sm:flex-row sm:items-center gap-3 sm:justify-between">
                            <div className="flex flex-col">
                                <span className="text-2xl font-bold text-gray-900 dark:text-gray-0">Inter</span>
                                <span className="text-sm text-gray-600 dark:text-gray-400">Used for the website, the panels and the logo name. Source Code Pro for code.</span>
                            </div>
                            <a href="https://fonts.google.com/specimen/Inter" className={link}>
                                Get Inter on Google Fonts →
                            </a>
                        </div>
                    </section>

                    <section className={card}>
                        <div className="flex flex-col gap-1">
                            <h2 className="text-xl font-bold text-gray-900 dark:text-gray-0 m-0">Screenshots</h2>
                            <p className="text-sm text-gray-600 dark:text-gray-400 m-0">Product screenshots for articles and reviews, in light and dark mode.</p>
                        </div>
                        <div className="grid gap-4 sm:grid-cols-2">
                            {SCREENSHOTS.map((s) => {
                                const dark = s.noDark ? undefined : s.light.replace(/\.png$/, "_dark.png");
                                return (
                                    <div key={s.name} className="rounded-xl border border-gray-200 dark:border-gray-700 overflow-hidden flex flex-col">
                                        <a href={s.light} className="block aspect-[16/10] bg-gray-50 dark:bg-gray-800 overflow-hidden">
                                            <img src={s.light} alt={s.name} loading="lazy" className={clsx("w-full h-full object-cover object-left-top", dark && "dark:hidden")} />
                                            {dark && <img src={dark} alt={s.name} loading="lazy" className="w-full h-full object-cover object-left-top hidden dark:block" />}
                                        </a>
                                        <div className="p-3 flex flex-wrap items-center justify-between gap-2">
                                            <span className="text-sm font-semibold text-gray-900 dark:text-gray-100">{s.name}</span>
                                            <div className="flex gap-2">
                                                <a href={s.light} download className={smallBtn}>
                                                    {dark ? "Light PNG" : "PNG"}
                                                </a>
                                                {dark && (
                                                    <a href={dark} download className={smallBtn}>
                                                        Dark PNG
                                                    </a>
                                                )}
                                            </div>
                                        </div>
                                    </div>
                                );
                            })}
                        </div>
                        <p className="text-sm text-gray-600 dark:text-gray-400 m-0">
                            More screenshots of every page are in the{" "}
                            <a href="/docs/" className={link}>
                                documentation
                            </a>
                            , or try it yourself on the{" "}
                            <a href="/demo/" className={link}>
                                live demo
                            </a>
                            .
                        </p>
                    </section>

                    <section className="grid gap-6 lg:grid-cols-2">
                        <div className={card}>
                            <h2 className="text-xl font-bold text-gray-900 dark:text-gray-0 m-0">Please do</h2>
                            <ul className="m-0 pl-5 list-disc flex flex-col gap-2 text-sm text-gray-700 dark:text-gray-300">
                                <li>Use the logo to link to OpenPanel, in reviews, tutorials and comparisons</li>
                                <li>Show it in hosting plans and integrations that run on OpenPanel</li>
                                <li>Leave clear space around it, about half the icon's width</li>
                            </ul>
                        </div>
                        <div className={card}>
                            <h2 className="text-xl font-bold text-gray-900 dark:text-gray-0 m-0">Please don't</h2>
                            <ul className="m-0 pl-5 list-disc flex flex-col gap-2 text-sm text-gray-700 dark:text-gray-300">
                                <li>Stretch, rotate, outline or add effects to the logo</li>
                                <li>Use it as your own product's logo or in a way that suggests a partnership we haven't agreed on</li>
                                <li>Put it on a background where it's hard to see</li>
                            </ul>
                        </div>
                    </section>

                    <p className="text-sm text-gray-600 dark:text-gray-400 m-0">
                        Need something else, like a different format or a partner badge? Write to{" "}
                        <a href="mailto:info@openpanel.com" className={link}>
                            info@openpanel.com
                        </a>
                        .
                    </p>
                </div>
                <BlogFooter />
            </div>
        </CommonLayout>
    );
};

export default Assets;
