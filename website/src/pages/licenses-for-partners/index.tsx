import React, { useState } from "react";
import clsx from "clsx";
import Head from "@docusaurus/Head";
import { CommonLayout } from "@site/src/refine-theme/common-layout";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";

const ORDER = "https://my.openpanel.com/index.php?rp=/store/partners/noc";
const TICKET = "https://my.openpanel.com/submitticket.php?step=2&deptid=1";
// WHMCS product 5 is the NOC package, its "IP range" option values go 1 = /28 up to 15 = /14
const orderUrl = (bits: number) =>
    `https://my.openpanel.com/cart.php?a=add&pid=5&billingcycle=monthly&configoption%5B1%5D=${29 - bits}`;
const PER_IP = 10;
// ranges bigger than /14 are quoted individually
const LARGEST_LISTED = 14;

const mask = (bits: number) => {
    const m = bits === 0 ? 0 : (~0 << (32 - bits)) >>> 0;
    return [24, 16, 8, 0].map((s) => (m >>> s) & 255).join(".");
};
const hosts = (bits: number) => 2 ** (32 - bits) - 2;
const eur = (n: number) =>
    "€" + n.toLocaleString("de-DE", { minimumFractionDigits: 2, maximumFractionDigits: 2 });
const num = (n: number) => n.toLocaleString("en-US");

const ROWS = Array.from({ length: 28 - 8 + 1 }, (_, i) => 28 - i).map((bits) => ({
    bits,
    mask: mask(bits),
    hosts: hosts(bits),
    price: bits >= LARGEST_LISTED ? hosts(bits) * PER_IP : undefined,
}));

const card = "rounded-2xl border border-gray-300 dark:border-gray-700 p-5 sm:p-6";
const primaryBtn =
    "rounded-full px-5 py-2.5 text-sm font-semibold bg-refine-blue dark:bg-refine-cyan-alt text-gray-0 dark:text-gray-900 hover:brightness-110 no-underline";
const secondaryBtn =
    "rounded-full px-5 py-2.5 text-sm font-semibold border border-gray-300 dark:border-gray-700 text-gray-900 dark:text-gray-100 hover:bg-gray-100 dark:hover:bg-gray-800 no-underline";

const NetworkTablePage: React.FC = () => {
    const [bits, setBits] = useState(24);
    const row = ROWS.find((r) => r.bits === bits)!;

    return (
        <>
            <Head>
                <title>NOC Partners Pricing | OpenPanel Enterprise</title>
                <meta property="og:title" content="NOC Partners Pricing | OpenPanel Enterprise" />
                <meta property="og:image:width" content="1200" />
                <meta property="og:image:height" content="630" />
                <meta property="og:image:type" content="image/png" />
                <html data-page="network-table" data-customized="true" />
            </Head>

            <div className="refine-prose">
                <CommonHeader hasSticky={true} />

                <div className="not-prose flex flex-col gap-12 pt-8 lg:pt-16 pb-32 max-w-[1000px] w-full mx-auto px-4">
                    <section className="flex flex-col gap-5">
                        <span className="self-start rounded-full px-3 py-1 text-xs font-semibold bg-refine-blue/10 text-refine-blue dark:bg-refine-cyan-alt/10 dark:text-refine-cyan-alt">
                            NOC partner program
                        </span>
                        <h1 className="text-3xl sm:text-5xl font-bold text-gray-900 dark:text-gray-0 m-0 max-w-[800px]">
                            OpenPanel Enterprise for your whole IP range
                        </h1>
                        <p className="text-base sm:text-lg text-gray-600 dark:text-gray-400 m-0 max-w-[720px]">
                            Own a /24 or any other range? License every IP in it with one NOC partner package, priced by the
                            network prefix length.
                        </p>
                        <div className="flex flex-wrap gap-3">
                            <a href={row.price !== undefined ? orderUrl(row.bits) : ORDER} className={primaryBtn}>
                                Order NOC package
                            </a>
                            <a href="#pricing" className={secondaryBtn}>
                                See all prices
                            </a>
                        </div>
                    </section>

                    <section className="grid gap-5 lg:grid-cols-[1fr_1.2fr] items-stretch">
                        <div className={clsx(card, "flex flex-col gap-4")}>
                            <h2 className="text-xl font-bold text-gray-900 dark:text-gray-0 m-0">How the price works</h2>
                            <div className="flex items-baseline gap-2">
                                <span className="text-4xl font-bold text-gray-900 dark:text-gray-0">{eur(PER_IP)}</span>
                                <span className="text-sm text-gray-600 dark:text-gray-400">per usable IP, per month</span>
                            </div>
                            <ul className="m-0 p-0 list-none flex flex-col gap-2 text-sm text-gray-800 dark:text-gray-200">
                                {[
                                    "One package covers every usable IP in the range",
                                    "Usable IPs are the range size minus network and broadcast",
                                    `Ranges larger than /${LARGEST_LISTED} are quoted individually`,
                                ].map((t) => (
                                    <li key={t} className="flex gap-2">
                                        <span className="text-refine-blue dark:text-refine-cyan-alt font-bold">✓</span>
                                        {t}
                                    </li>
                                ))}
                            </ul>
                        </div>

                        <div className={clsx(card, "flex flex-col gap-4")}>
                            <h2 className="text-xl font-bold text-gray-900 dark:text-gray-0 m-0">Your range</h2>
                            <label className="flex flex-col gap-1.5 text-sm font-semibold text-gray-700 dark:text-gray-300">
                                Prefix length
                                <select
                                    value={bits}
                                    onChange={(e) => setBits(+e.target.value)}
                                    className="rounded-lg border border-gray-300 dark:border-gray-700 bg-gray-0 dark:bg-gray-900 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm"
                                >
                                    {ROWS.map((r) => (
                                        <option key={r.bits} value={r.bits}>
                                            /{r.bits} ({num(r.hosts)} usable IPs)
                                        </option>
                                    ))}
                                </select>
                            </label>
                            <dl className="m-0 grid grid-cols-2 gap-3">
                                {[
                                    ["Subnet mask", row.mask],
                                    ["Usable IPs", num(row.hosts)],
                                    ["Monthly", row.price !== undefined ? eur(row.price) : "Quoted"],
                                    ["Yearly", row.price !== undefined ? eur(row.price * 12) : "Quoted"],
                                ].map(([k, v]) => (
                                    <div key={k} className="rounded-xl bg-gray-50 dark:bg-gray-800/60 p-3">
                                        <dt className="text-xs text-gray-500 dark:text-gray-400">{k}</dt>
                                        <dd className="m-0 text-lg font-bold text-gray-900 dark:text-gray-0 font-mono">{v}</dd>
                                    </div>
                                ))}
                            </dl>
                            <a href={row.price !== undefined ? orderUrl(row.bits) : TICKET} className={clsx(primaryBtn, "self-start")}>
                                {row.price !== undefined ? `Order /${row.bits}` : "Ask for a quote"}
                            </a>
                        </div>
                    </section>

                    <section id="pricing" className="flex flex-col gap-4 scroll-mt-28">
                        <h2 className="text-2xl sm:text-3xl font-bold text-gray-900 dark:text-gray-0 m-0">All prices</h2>
                        <div className="overflow-x-auto rounded-2xl border border-gray-300 dark:border-gray-700">
                            <table className="w-full text-left text-sm m-0 border-collapse" style={{ display: "table" }}>
                                <thead>
                                    <tr className="bg-gray-50 dark:bg-gray-800 text-gray-700 dark:text-gray-300">
                                        <th className="!border-0 px-4 py-3 font-semibold">Prefix</th>
                                        <th className="!border-0 px-4 py-3 font-semibold">Subnet mask</th>
                                        <th className="!border-0 px-4 py-3 font-semibold text-right">Usable IPs</th>
                                        <th className="!border-0 px-4 py-3 font-semibold text-right">EUR / month</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {ROWS.map((r) => (
                                        <tr
                                            key={r.bits}
                                            onClick={() => setBits(r.bits)}
                                            style={r.bits === bits ? { backgroundColor: "rgba(56, 189, 248, 0.18)" } : undefined}
                                            className={clsx(
                                                "!border-0 border-t border-solid !border-t-gray-200 dark:!border-t-gray-700 cursor-pointer",
                                                r.bits !== bits && "hover:bg-gray-50 dark:hover:bg-gray-800/60",
                                            )}
                                        >
                                            <td className="!border-0 px-4 py-2.5 font-mono font-semibold text-gray-900 dark:text-gray-100">/{r.bits}</td>
                                            <td className="!border-0 px-4 py-2.5 font-mono text-gray-600 dark:text-gray-400">{r.mask}</td>
                                            <td className="!border-0 px-4 py-2.5 font-mono text-right text-gray-800 dark:text-gray-200">{num(r.hosts)}</td>
                                            <td className="!border-0 px-4 py-2.5 font-mono text-right text-gray-900 dark:text-gray-100">
                                                {r.price !== undefined ? (
                                                    eur(r.price)
                                                ) : (
                                                    <a href={TICKET} className="text-refine-react-light-link dark:text-refine-react-dark-link hover:underline font-sans">
                                                        Contact us
                                                    </a>
                                                )}
                                            </td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </div>
                        <p className="text-xs text-gray-500 dark:text-gray-400 m-0">Click a row to see it in the calculator above.</p>
                    </section>

                    <section className={clsx(card, "flex flex-col sm:flex-row gap-4 sm:items-center sm:justify-between")}>
                        <div className="flex flex-col gap-1">
                            <h2 className="text-xl font-bold text-gray-900 dark:text-gray-0 m-0">Need a single server instead?</h2>
                            <p className="text-sm text-gray-600 dark:text-gray-400 m-0">
                                See the regular{" "}
                                <a href="/enterprise/" className="text-refine-react-light-link dark:text-refine-react-dark-link hover:underline">
                                    Enterprise edition
                                </a>{" "}
                                or ask us about your network through a{" "}
                                <a href={TICKET} className="text-refine-react-light-link dark:text-refine-react-dark-link hover:underline">
                                    support ticket
                                </a>
                                .
                            </p>
                        </div>
                        <a href={row.price !== undefined ? orderUrl(row.bits) : ORDER} className={clsx(primaryBtn, "self-start sm:self-auto whitespace-nowrap")}>
                            Order NOC package
                        </a>
                    </section>
                </div>

                <BlogFooter />
            </div>
        </>
    );
};

export default function NetworkTable() {
    return (
        <CommonLayout
            description="OpenPanel Enterprise pricing for NOC partners with their own IP ranges, priced by network prefix length."
            image="/img/og/licenses-for-partners.png"
        >
            <NetworkTablePage />
        </CommonLayout>
    );
}
