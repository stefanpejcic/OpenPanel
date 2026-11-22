import React from "react";
import clsx from "clsx";
import Head from "@docusaurus/Head";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";

const TICKET = "https://my.openpanel.com/submitticket.php?step=2&deptid=1";
const LICENSE = "https://my.openpanel.com/index.php?rp=/store/openpanel/enterprise-license";

const link = "text-refine-react-light-link dark:text-refine-react-dark-link hover:underline";
const card = "rounded-2xl border border-gray-300 dark:border-gray-700 p-5 sm:p-6";
const h2 = "text-2xl sm:text-3xl font-bold text-gray-900 dark:text-gray-0 m-0";
const muted = "text-sm text-gray-600 dark:text-gray-400 m-0";
const primaryBtn =
    "rounded-full px-5 py-2.5 text-sm font-semibold bg-refine-blue dark:bg-refine-cyan-alt text-gray-0 dark:text-gray-900 hover:brightness-110 no-underline";
const secondaryBtn =
    "rounded-full px-5 py-2.5 text-sm font-semibold border border-gray-300 dark:border-gray-700 text-gray-900 dark:text-gray-100 hover:bg-gray-100 dark:hover:bg-gray-800 no-underline";

const PANELS = [
    { name: "cPanel / WHM", note: "All versions with root or WHM access", href: "/cpanel-alternative/" },
    { name: "Plesk", note: "Linux servers only", href: "/plesk-alternative/" },
    { name: "DirectAdmin", note: "Admin or root access required", href: "/directadmin-alternative/" },
    { name: "CyberPanel", note: "Root access required", href: "/cyberpanel-alternative/" },
    { name: "OpenPanel", note: "Community or Enterprise edition, root access required" },
];

const NOT_SUPPORTED = [
    "Custom server stacks (plain LAMP/LEMP, Docker setups, bare VPS without a control panel)",
    "Windows / IIS servers, including Plesk for Windows",
    "Website builders (Wix, Squarespace, Weebly and similar)",
    "Any control panel not listed above",
];

const MIGRATED = [
    "User accounts and their domains, subdomains, addon and parked/alias domains",
    "Website files",
    "MySQL / MariaDB databases and database users",
    "Email accounts, mailboxes, forwarders and autoresponders",
    "DNS zones and records",
    "Cron jobs",
    "FTP accounts",
    "SSL certificates (existing certificates are copied, or new Let's Encrypt certificates are issued once DNS points to the new server)",
];

const NOT_INCLUDED = [
    "Server-level custom configuration (custom Apache/Nginx rules, custom PHP extensions or compiled modules, kernel or firewall tuning)",
    "Software or services not managed by the source panel",
    "Billing system data (WHMCS, Blesta, etc.)",
    "Domain name transfers between registrars",
    "Changing DNS records or nameservers at your registrar, this is done by you once you have verified the migrated sites",
    "Merging several source accounts into one account, or splitting one account into several",
    "Fixing issues that already existed on the source server (broken sites, outdated or incompatible application code)",
];

const STEPS: { title: string; body: React.ReactNode }[] = [
    {
        title: "Open a ticket",
        body: (
            <>
                Submit a{" "}
                <a href={TICKET} className={link}>
                    migration request ticket
                </a>{" "}
                from the account that owns the annual license.
            </>
        ),
    },
    {
        title: "Provide details",
        body: (
            <ul className="m-0 mt-1 pl-5 list-disc flex flex-col gap-1">
                <li>Your license key and the destination server IP</li>
                <li>Source panel name and version</li>
                <li>Number of accounts to migrate and approximate total disk usage</li>
                <li>Login credentials for the source server: root SSH access (preferred), or admin login for the panel</li>
                <li>Root SSH access to the destination OpenPanel server</li>
            </ul>
        ),
    },
    {
        title: "Review",
        body: "We check access and size, confirm the number of accounts and, if the migration exceeds 20 accounts or 500 GB, send an invoice for the difference.",
    },
    { title: "Scheduling", body: "We agree on a migration window with you." },
    {
        title: "Migration",
        body: "We transfer all accounts to the OpenPanel server. Your source server keeps running and serving traffic during this time.",
    },
    {
        title: "Verification",
        body: "You test the migrated websites and email, for example by using a hosts file entry or the temporary domain.",
    },
    { title: "DNS switch", body: "You point your domains or nameservers to the new server." },
    {
        title: "Final sync",
        body: "On request, we run one final sync of files, databases and mailboxes changed since the initial migration.",
    },
    { title: "Closing", body: "We close the ticket and delete all credentials you provided." },
];

const RESPONSIBILITIES = [
    "Keep a full, current backup of the source server before and during the migration. OpenPanel LLC is not liable for data loss on the source or destination server.",
    "Make sure the destination server has enough disk space, memory and CPU for all migrated accounts.",
    "Lower the TTL of your DNS records at least 24 hours before the DNS switch to minimize downtime.",
    "Keep the source server online and accessible until the migration and final sync are complete.",
    "Change all passwords (server root, panel admin, SSH) after the migration is complete.",
];

const Check = ({ ok }: { ok: boolean }) => (
    <span
        aria-hidden
        className={clsx(
            "shrink-0 mt-0.5 w-5 h-5 rounded-full flex items-center justify-center text-[11px] font-bold",
            ok
                ? "bg-emerald-100 text-emerald-700 dark:bg-transparent dark:border dark:border-emerald-400/60 dark:text-emerald-300"
                : "bg-gray-100 text-gray-500 dark:bg-transparent dark:border dark:border-gray-600 dark:text-gray-400",
        )}
    >
        {ok ? "✓" : "✕"}
    </span>
);

const ItemList = ({ items, ok }: { items: string[]; ok: boolean }) => (
    <ul className="m-0 p-0 list-none flex flex-col gap-2.5">
        {items.map((i) => (
            <li key={i} className="flex gap-2.5 text-sm text-gray-800 dark:text-gray-200">
                <Check ok={ok} />
                <span>{i}</span>
            </li>
        ))}
    </ul>
);

const Migrations: React.FC = () => {
    return (
        <CommonLayout
            description="Free migration from cPanel, Plesk, DirectAdmin and CyberPanel to OpenPanel for annual Enterprise license holders, up to 20 accounts and 500 GB included."
            image="/img/og/migrations.png"
        >
            <Head>
                <title>Free Migration to OpenPanel | OpenPanel</title>
                <meta property="og:title" content="Free Migration to OpenPanel | OpenPanel" />
                <meta property="og:image:width" content="1200" />
                <meta property="og:image:height" content="630" />
                <meta property="og:image:type" content="image/png" />
                <html data-page="migrations" data-customized="true" />
            </Head>
            <div className="refine-prose">
                <CommonHeader hasSticky={true} />

                <div className="not-prose flex flex-col gap-14 pt-8 lg:pt-16 pb-32 max-w-[928px] landing-lg:max-w-[1232px] w-full mx-auto px-4">
                    <section className="flex flex-col gap-5">
                        <span className="self-start rounded-full px-3 py-1 text-xs font-semibold bg-refine-blue/10 text-refine-blue dark:bg-refine-cyan-alt/10 dark:text-refine-cyan-alt">
                            Included with annual Enterprise licenses
                        </span>
                        <h1 className="text-3xl sm:text-5xl font-bold text-gray-900 dark:text-gray-0 m-0 max-w-[780px]">
                            Free migration to OpenPanel
                        </h1>
                        <p className="text-base sm:text-lg text-gray-600 dark:text-gray-400 m-0 max-w-[720px]">
                            Send us access to your current server and our team moves your websites, databases, email and DNS
                            to OpenPanel at no extra cost. Your old server keeps serving traffic until you switch.
                        </p>
                        <div className="flex flex-wrap gap-3">
                            <a href={TICKET} className={primaryBtn}>
                                Request a migration
                            </a>
                            <a href={LICENSE} className={secondaryBtn}>
                                Get an annual license
                            </a>
                        </div>
                        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 mt-2">
                            {[
                                ["20", "accounts included"],
                                ["500 GB", "of data included"],
                                ["€10", "per account above 20"],
                                ["3 days", "usual time to start"],
                            ].map(([big, small]) => (
                                <div key={small} className="rounded-xl border border-gray-200 dark:border-gray-700 p-4">
                                    <div className="text-2xl font-bold text-gray-900 dark:text-gray-0">{big}</div>
                                    <div className="text-sm text-gray-600 dark:text-gray-400">{small}</div>
                                </div>
                            ))}
                        </div>
                        <p className="text-xs text-gray-500 dark:text-gray-500 m-0">Effective date: October 2026</p>
                    </section>

                    <section id="panels" className="flex flex-col gap-5 scroll-mt-28">
                        <div className="flex flex-col gap-2">
                            <h2 className={h2}>Supported source panels</h2>
                            <p className={muted}>The source server has to run one of these control panels.</p>
                        </div>
                        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
                            {PANELS.map((p) => {
                                const inner = (
                                    <>
                                        <div className="shrink-0 w-10 h-10 rounded-lg bg-gray-100 dark:bg-gray-800 flex items-center justify-center text-base font-bold text-gray-700 dark:text-gray-200">
                                            {p.name[0]}
                                        </div>
                                        <div className="flex flex-col gap-0.5">
                                            <span className="text-sm font-semibold text-gray-900 dark:text-gray-100">{p.name}</span>
                                            <span className="text-sm text-gray-600 dark:text-gray-400">{p.note}</span>
                                        </div>
                                    </>
                                );
                                const cls = "flex gap-3 items-start rounded-xl border border-gray-200 dark:border-gray-700 p-4 no-underline";
                                return p.href ? (
                                    <a key={p.name} href={p.href} className={clsx(cls, "hover:border-gray-400 dark:hover:border-gray-500")}>
                                        {inner}
                                    </a>
                                ) : (
                                    <div key={p.name} className={cls}>
                                        {inner}
                                    </div>
                                );
                            })}
                        </div>
                        <div className="rounded-xl bg-gray-50 dark:bg-gray-800/60 p-4 flex flex-col gap-2.5">
                            <span className="text-sm font-semibold text-gray-900 dark:text-gray-100">Not supported</span>
                            <ItemList items={NOT_SUPPORTED} ok={false} />
                        </div>
                    </section>

                    <section id="eligibility" className="grid gap-5 lg:grid-cols-2 scroll-mt-28">
                        <div className={clsx(card, "flex flex-col gap-4")}>
                            <h2 className="text-xl font-bold text-gray-900 dark:text-gray-0 m-0">Who can get it</h2>
                            <p className={muted}>All of these have to be true:</p>
                            <ItemList
                                ok
                                items={[
                                    "You have an active yearly (annual billing) Enterprise license. Monthly licenses, trial licenses and the Community Edition are not eligible.",
                                    "The license is activated on the destination server, and OpenPanel is installed and reachable.",
                                    "The source server runs one of the supported control panels.",
                                    "The request is submitted by the license holder through our support system.",
                                ]}
                            />
                            <p className={muted}>One free migration is included per annual license.</p>
                        </div>
                        <div className={clsx(card, "flex flex-col gap-4")}>
                            <h2 className="text-xl font-bold text-gray-900 dark:text-gray-0 m-0">Limits & pricing</h2>
                            <dl className="m-0 grid grid-cols-[auto_1fr] gap-x-4 gap-y-3 text-sm">
                                <dt className="text-gray-600 dark:text-gray-400">Included free</dt>
                                <dd className="m-0 text-gray-900 dark:text-gray-100">
                                    Up to <b>20 accounts</b> and <b>500 GB</b> of total data per migration
                                </dd>
                                <dt className="text-gray-600 dark:text-gray-400">Additional accounts</dt>
                                <dd className="m-0 text-gray-900 dark:text-gray-100">
                                    <b>€10 per account</b> above 20
                                </dd>
                                <dt className="text-gray-600 dark:text-gray-400">Data above 500 GB</dt>
                                <dd className="m-0 text-gray-900 dark:text-gray-100">Quoted individually</dd>
                            </dl>
                            <p className={muted}>
                                Total data includes website files, databases and mailboxes across all migrated accounts.
                                Additional accounts and data are invoiced before the migration starts.
                            </p>
                            <p className={muted}>
                                An "account" is one hosting user on the source panel (a cPanel account, a Plesk subscription, a
                                DirectAdmin user, a CyberPanel website owner or an OpenPanel user), including all of its
                                domains, databases and mailboxes.
                            </p>
                        </div>
                    </section>

                    <section id="scope" className="grid gap-5 lg:grid-cols-2 scroll-mt-28">
                        <div className={clsx(card, "flex flex-col gap-4")}>
                            <h2 className="text-xl font-bold text-gray-900 dark:text-gray-0 m-0">What we migrate</h2>
                            <ItemList items={MIGRATED} ok />
                        </div>
                        <div className={clsx(card, "flex flex-col gap-4")}>
                            <h2 className="text-xl font-bold text-gray-900 dark:text-gray-0 m-0">What is not included</h2>
                            <ItemList items={NOT_INCLUDED} ok={false} />
                        </div>
                    </section>

                    <section id="procedure" className="flex flex-col gap-6 scroll-mt-28">
                        <div className="flex flex-col gap-2">
                            <h2 className={h2}>How it works</h2>
                            <p className={muted}>From the ticket to the final sync, your source server stays online the whole time.</p>
                        </div>
                        <ol className="m-0 p-0 list-none flex flex-col">
                            {STEPS.map((s, i) => (
                                <li key={s.title} className="flex gap-4">
                                    <div className="flex flex-col items-center">
                                        <span className="w-8 h-8 shrink-0 rounded-full bg-refine-blue dark:bg-refine-cyan-alt text-gray-0 dark:text-gray-900 text-sm font-bold flex items-center justify-center">
                                            {i + 1}
                                        </span>
                                        {i < STEPS.length - 1 && <span className="w-px flex-1 bg-gray-300 dark:bg-gray-700" />}
                                    </div>
                                    <div className={clsx("flex flex-col gap-1 pt-1", i < STEPS.length - 1 && "pb-6")}>
                                        <span className="text-base font-semibold text-gray-900 dark:text-gray-100">{s.title}</span>
                                        <div className="text-sm text-gray-600 dark:text-gray-400">{s.body}</div>
                                    </div>
                                </li>
                            ))}
                        </ol>
                    </section>

                    <section className="grid gap-5 lg:grid-cols-3">
                        <div className={clsx(card, "flex flex-col gap-4 lg:col-span-2")}>
                            <h2 className="text-xl font-bold text-gray-900 dark:text-gray-0 m-0">Your responsibilities</h2>
                            <ItemList items={RESPONSIBILITIES} ok />
                        </div>
                        <div className="flex flex-col gap-5">
                            <div className={clsx(card, "flex flex-col gap-2")}>
                                <h3 className="text-base font-bold text-gray-900 dark:text-gray-0 m-0">Credentials & security</h3>
                                <p className={muted}>
                                    Credentials are accepted <b>only through our support ticket system</b>, never send them by
                                    email, chat or social media. They are used only for the migration, are not shared with third
                                    parties, and are removed from the ticket once the migration is complete.
                                </p>
                            </div>
                            <div className={clsx(card, "flex flex-col gap-2")}>
                                <h3 className="text-base font-bold text-gray-900 dark:text-gray-0 m-0">Timeframe</h3>
                                <p className={muted}>
                                    Migrations usually start within 3 business days after all required information is received.
                                    Duration depends on the number of accounts and the amount of data.
                                </p>
                            </div>
                        </div>
                    </section>

                    <section className="flex flex-col gap-3">
                        <h2 className="text-xl font-bold text-gray-900 dark:text-gray-0 m-0">Right to refuse</h2>
                        <p className={muted}>
                            We may decline or stop a migration if the source server is not accessible, the data is corrupted or
                            incomplete, the content violates our terms of service, or the migration is not technically possible.
                            In that case we will explain the reason in the ticket.
                        </p>
                        <p className={muted}>
                            A migration does not change the{" "}
                            <a href="/refund-policy" className={link}>
                                Refund Policy
                            </a>
                            : the license is activated before migration and activated licenses are not refundable.
                        </p>
                    </section>

                    <section className={clsx(card, "flex flex-col sm:flex-row gap-4 sm:items-center sm:justify-between")}>
                        <div className="flex flex-col gap-1">
                            <h2 className="text-xl font-bold text-gray-900 dark:text-gray-0 m-0">Not sure your server qualifies?</h2>
                            <p className={muted}>
                                Ask us before purchasing through a{" "}
                                <a href={TICKET} className={link}>
                                    support ticket
                                </a>{" "}
                                or at{" "}
                                <a href="mailto:info@openpanel.com" className={link}>
                                    info@openpanel.com
                                </a>
                                .
                            </p>
                        </div>
                        <a href={TICKET} className={clsx(primaryBtn, "self-start sm:self-auto whitespace-nowrap")}>
                            Request a migration
                        </a>
                    </section>
                </div>
                <BlogFooter />
            </div>
        </CommonLayout>
    );
};

export default Migrations;
