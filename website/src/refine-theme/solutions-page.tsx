import Head from "@docusaurus/Head";
import Link from "@docusaurus/Link";
import clsx from "clsx";
import React, { useState } from "react";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";
import { LandingFooter } from "@site/src/refine-theme/landing-footer";
import { FaqSchema } from "@site/src/refine-theme/faq-schema";
import { EnterpriseGetInTouchButton } from "@site/src/refine-theme/enterprise-get-in-touch-button";

// Shared layout for the /solutions/* audience pages (hosting providers,
// agencies). Each page passes its own copy, the section structure stays the
// same so both pages read as one family.

export type SolutionChallenge = {
    title: string;
    problem: string;
    solvedBy: string[];
    product: string;
    description: string;
    instead: string;
    outcome: string;
    link?: string;
    // docs screenshot path without extension; "<image>.png" is the light
    // variant and "<image>_dark.png" the dark one
    image?: string;
    imageAlt?: string;
};

export type SolutionStat = { value: string; label: string };
export type SolutionColumn = { title: string; text: string };
export type SolutionFaq = {
    question: string;
    // plain text, also used for the FAQPage structured data
    answer: string;
    // optional rich version (links) shown on the page instead of `answer`
    answerNode?: React.ReactNode;
};
export type SolutionCompareRow = { feature: string; typical: string };
export type SolutionGroup = {
    title: string;
    subtitle: string;
    items: SolutionChallenge[];
};

export type SolutionPageProps = {
    slug: string;
    metaTitle: string;
    metaDescription: string;
    eyebrow: string;
    heroTitle: React.ReactNode;
    heroText: React.ReactNode;
    eventPrefix: string;
    stats: SolutionStat[];
    groups: SolutionGroup[];
    columnsTitle: string;
    columns: SolutionColumn[];
    compareTitle: string;
    compareSubtitle: string;
    compareRows: SolutionCompareRow[];
    extra?: React.ReactNode;
    faqTitle: string;
    faq: SolutionFaq[];
    ctaTitle: string;
    ctaText: string;
};

const sectionWidth = clsx(
    "mx-auto",
    "w-full",
    "px-4 landing-sm:px-0",
    "max-w-[592px]",
    "landing-sm:max-w-[656px]",
    "landing-md:max-w-[896px]",
    "landing-lg:max-w-[1200px]",
);

const cardBase = clsx(
    "dark:bg-landing-noise",
    "dark:bg-gray-800 bg-gray-50",
    "rounded-2xl",
);

const accent = clsx(
    "dark:text-refine-cyan-alt dark:drop-shadow-[0_0_30px_rgba(71,235,235,0.25)]",
    "text-refine-blue drop-shadow-[0_0_30px_rgba(51,51,255,0.3)]",
);

export const SolutionAccent = ({ children }: React.PropsWithChildren) => (
    <span className={clsx("font-semibold", accent)}>{children}</span>
);

const SectionHeading = ({
    title,
    subtitle,
    id,
}: {
    title: string;
    subtitle?: string;
    id?: string;
}) => (
    <div className="pl-0 landing-sm:pl-6 landing-md:pl-10">
        <h2
            id={id}
            className={clsx(
                "font-semibold",
                "text-2xl landing-sm:text-[32px] landing-sm:leading-[40px]",
                "dark:text-gray-0 text-gray-900",
            )}
        >
            {title}
        </h2>
        {subtitle && (
            <p
                className={clsx(
                    "mt-3 max-w-[640px]",
                    "text-base",
                    "dark:text-gray-400 text-gray-600",
                )}
            >
                {subtitle}
            </p>
        )}
    </div>
);

const CheckIcon = ({ className }: { className?: string }) => (
    <svg
        viewBox="0 0 16 16"
        width="16"
        height="16"
        fill="none"
        aria-hidden="true"
        className={className}
    >
        <circle cx="8" cy="8" r="8" className="fill-current opacity-20" />
        <path
            d="M4.75 8.25 7 10.5l4.25-5"
            stroke="currentColor"
            strokeWidth="1.6"
            strokeLinecap="round"
            strokeLinejoin="round"
        />
    </svg>
);

const PlusIcon = ({ className }: { className?: string }) => (
    <svg
        viewBox="0 0 16 16"
        width="16"
        height="16"
        fill="none"
        aria-hidden="true"
        className={className}
    >
        <circle cx="8" cy="8" r="8" className="fill-current opacity-20" />
        <path
            d="M8 4.5v7M4.5 8h7"
            stroke="currentColor"
            strokeWidth="1.6"
            strokeLinecap="round"
        />
    </svg>
);

// Hero visual: the add-on licence stack a typical panel needs, next to the
// single OpenPanel licence that already contains it. On load the typical
// stack builds up one licence at a time, then the whole OpenPanel column
// lands at once to make the contrast obvious.
const STEP = 0.35; // seconds between typical-stack rows
const FIRST = 0.3;

const licenseStackCss = `
@keyframes lsIn { from { opacity: 0; transform: translateY(6px); } to { opacity: 1; transform: none; } }
@keyframes lsPop { 0% { opacity: 0; transform: scale(.96); } 60% { opacity: 1; transform: scale(1.015); } 100% { opacity: 1; transform: none; } }
@keyframes lsGlow { 0% { box-shadow: 0 0 0 0 rgba(16,185,129,.0); } 40% { box-shadow: 0 0 0 6px rgba(16,185,129,.18); } 100% { box-shadow: 0 0 0 0 rgba(16,185,129,0); } }
.ls-in { opacity: 0; animation: lsIn .45s ease-out forwards; }
.ls-pop { opacity: 0; animation: lsPop .6s cubic-bezier(.2,.8,.2,1) forwards; }
.ls-glow { opacity: 0; animation: lsPop .6s cubic-bezier(.2,.8,.2,1) forwards, lsGlow 1.2s ease-out forwards; }
@media (prefers-reduced-motion: reduce) {
  .ls-in, .ls-pop, .ls-glow { animation: none !important; opacity: 1 !important; transform: none !important; }
}
`;

const delay = (s: number) => ({ animationDelay: `${s.toFixed(2)}s` });

const LicenseStackVisual = ({ rows }: { rows: SolutionCompareRow[] }) => {
    const items = rows.slice(0, 6);
    // control panel row + one row per add-on, then the "N licences" note
    const leftDone = FIRST + (items.length + 1) * STEP;
    const rightAt = leftDone + 0.5;
    return (
        <div
            className={clsx(
                cardBase,
                "p-4 landing-sm:p-6",
                "grid grid-cols-1 landing-sm:grid-cols-2 gap-6 landing-sm:gap-4",
                "text-xs landing-sm:text-sm",
            )}
            aria-label="A typical hosting stack needs a separate licence for each layer, OpenPanel includes them in one licence"
        >
            <style>{licenseStackCss}</style>
            <div className="flex flex-col gap-2">
                <div
                    className={clsx(
                        "font-semibold uppercase tracking-wide text-[11px]",
                        "dark:text-gray-500 text-gray-500",
                    )}
                >
                    Typical stack
                </div>
                <div
                    className={clsx(
                        "ls-in",
                        "rounded-lg px-3 py-2",
                        "dark:bg-gray-700 bg-gray-200",
                        "dark:text-gray-200 text-gray-800",
                        "font-medium",
                    )}
                    style={delay(FIRST)}
                >
                    Control panel licence
                </div>
                {items.map((r, i) => (
                    <div
                        key={r.feature}
                        className={clsx(
                            "ls-in",
                            "flex items-center justify-between gap-2",
                            "rounded-lg px-3 py-2",
                            "border border-dashed",
                            "dark:border-gray-600 border-gray-300",
                            "dark:text-gray-400 text-gray-600",
                        )}
                        style={delay(FIRST + (i + 1) * STEP)}
                    >
                        <span className="truncate">{r.feature}</span>
                        <PlusIcon className="shrink-0 text-orange-500" />
                    </div>
                ))}
                <div
                    className={clsx(
                        "ls-in",
                        "mt-1 text-center text-[11px]",
                        "dark:text-orange-400 text-orange-600",
                    )}
                    style={delay(leftDone)}
                >
                    {items.length + 1} licences, {items.length + 1} renewals
                </div>
            </div>
            <div className="flex flex-col gap-2">
                <div
                    className={clsx(
                        "ls-in",
                        "font-semibold uppercase tracking-wide text-[11px]",
                        "dark:text-refine-cyan-alt text-refine-blue",
                    )}
                    style={delay(rightAt - 0.15)}
                >
                    OpenPanel Enterprise
                </div>
                <div
                    className={clsx(
                        "ls-glow",
                        "flex-1 flex flex-col gap-2",
                        "rounded-xl p-3",
                        "border",
                        "dark:border-refine-cyan-alt/40 border-refine-blue/30",
                        "dark:bg-refine-cyan-alt/5 bg-refine-blue/5",
                    )}
                    style={delay(rightAt)}
                >
                    <div
                        className={clsx(
                            "rounded-lg px-3 py-2",
                            "dark:bg-gray-700 bg-white",
                            "dark:text-gray-0 text-gray-900",
                            "font-medium",
                        )}
                    >
                        Control panel
                    </div>
                    {items.map((r) => (
                        <div
                            key={r.feature}
                            className={clsx(
                                "flex items-center justify-between gap-2",
                                "rounded-lg px-3 py-2",
                                "dark:bg-gray-700/60 bg-white",
                                "dark:text-gray-200 text-gray-800",
                            )}
                        >
                            <span className="truncate">{r.feature}</span>
                            <CheckIcon className="shrink-0 text-emerald-500" />
                        </div>
                    ))}
                </div>
                <div
                    className={clsx(
                        "ls-in",
                        "mt-1 text-center text-[11px] font-semibold",
                        "dark:text-gray-0 text-gray-900",
                    )}
                    style={delay(rightAt + 0.4)}
                >
                    1 licence · €14.95/month
                </div>
            </div>
        </div>
    );
};

const Hero = (p: SolutionPageProps) => (
    <div
        className={clsx(
            sectionWidth,
            "not-prose",
            "grid grid-cols-1 landing-md:grid-cols-2",
            "gap-10 landing-md:gap-12",
            "items-center",
            "mt-4 landing-sm:mt-8",
        )}
    >
        <div className="flex flex-col landing-sm:pl-6 landing-md:pl-10">
            <div
                className={clsx(
                    "text-xs font-semibold uppercase tracking-wider",
                    "dark:text-gray-500 text-gray-500",
                )}
            >
                {p.eyebrow}
            </div>
            <h1
                className={clsx(
                    "mt-3",
                    "text-[32px] leading-[40px] landing-sm:text-[48px] landing-sm:leading-[60px]",
                    "tracking-tight",
                    "dark:text-gray-0 text-gray-900",
                )}
            >
                {p.heroTitle}
            </h1>
            <p
                className={clsx(
                    "mt-6 max-w-[480px]",
                    "dark:text-gray-400 text-gray-600",
                )}
            >
                {p.heroText}
            </p>
            <div
                className={clsx(
                    "mt-8",
                    "flex flex-col landing-sm:flex-row",
                    "items-start landing-sm:items-center",
                    "gap-4 landing-sm:gap-6",
                )}
            >
                <EnterpriseGetInTouchButton
                    label="Start free 30-day trial"
                    href="/trial"
                    eventName={`trial_click_hero_${p.eventPrefix}`}
                />
                <Link
                    to="#stack"
                    className={clsx(
                        "text-sm font-semibold",
                        "dark:text-gray-0 text-gray-900",
                        "underline underline-offset-4",
                    )}
                >
                    See what's built in →
                </Link>
            </div>
            <p className="mt-3 text-xs dark:text-gray-500 text-gray-500">
                No credit card required · Cancel anytime
            </p>
        </div>
        <LicenseStackVisual rows={p.compareRows} />
    </div>
);

const Stats = ({ stats }: { stats: SolutionStat[] }) => (
    <div
        className={clsx(
            sectionWidth,
            "not-prose",
            "grid grid-cols-2 landing-md:grid-cols-4",
            "gap-4",
        )}
    >
        {stats.map((s) => (
            <div key={s.label} className={clsx(cardBase, "p-5 landing-sm:p-6")}>
                <div
                    className={clsx(
                        "text-[28px] leading-9 landing-sm:text-[40px] landing-sm:leading-[48px]",
                        "font-semibold tracking-tight",
                        accent,
                    )}
                >
                    {s.value}
                </div>
                <div className="mt-2 text-sm dark:text-gray-400 text-gray-600">
                    {s.label}
                </div>
            </div>
        ))}
    </div>
);

// Docs screenshot in a small window frame, swapping the light/dark variant
// with the site theme via CSS so there is no flash on first paint.
const ScreenshotFrame = ({ src, alt }: { src: string; alt: string }) => (
    <div
        className={clsx(
            "overflow-hidden rounded-xl",
            "border dark:border-gray-700 border-gray-200",
            "dark:bg-gray-900 bg-white",
            "shadow-sm",
        )}
    >
        <div
            className={clsx(
                "flex items-center gap-1.5 px-3 py-2",
                "border-b dark:border-gray-700 border-gray-200",
            )}
            aria-hidden="true"
        >
            <span className="h-2 w-2 rounded-full dark:bg-gray-600 bg-gray-300" />
            <span className="h-2 w-2 rounded-full dark:bg-gray-600 bg-gray-300" />
            <span className="h-2 w-2 rounded-full dark:bg-gray-600 bg-gray-300" />
        </div>
        <div className="aspect-[16/9] overflow-hidden">
            <img
                src={`${src}.png`}
                alt={alt}
                loading="lazy"
                className="block dark:hidden w-full h-full object-cover object-left-top"
            />
            <img
                src={`${src}_dark.png`}
                alt={alt}
                loading="lazy"
                className="hidden dark:block w-full h-full object-cover object-left-top"
            />
        </div>
    </div>
);

const ChallengeCard = ({
    item,
    index,
}: {
    item: SolutionChallenge;
    index: number;
}) => (
    <div
        className={clsx(
            cardBase,
            "flex flex-col",
            "p-6 landing-sm:p-8",
            "h-full",
        )}
    >
        <div className="flex items-center gap-3">
            <span
                className={clsx(
                    "font-mono text-sm font-semibold",
                    "dark:text-refine-cyan-alt text-refine-blue",
                )}
            >
                {String(index).padStart(2, "0")}
            </span>
            <span className="h-px flex-1 dark:bg-gray-700 bg-gray-200" />
        </div>
        <h3
            className={clsx(
                "mt-4",
                "text-lg landing-sm:text-xl font-semibold",
                "dark:text-gray-0 text-gray-900",
            )}
        >
            {item.title}
        </h3>
        <p className="mt-2 text-sm dark:text-gray-400 text-gray-600">
            {item.problem}
        </p>

        <div className="mt-5 flex flex-wrap items-center gap-2">
            <span className="text-xs dark:text-gray-500 text-gray-500">
                Solved by
            </span>
            {item.solvedBy.map((b) => (
                <span
                    key={b}
                    className={clsx(
                        "text-xs font-semibold",
                        "rounded-full px-2.5 py-1",
                        "dark:bg-refine-cyan-alt/10 bg-refine-blue/10",
                        "dark:text-refine-cyan-alt text-refine-blue",
                    )}
                >
                    {b}
                </span>
            ))}
        </div>

        <div
            className={clsx(
                "mt-5 text-base font-semibold",
                "dark:text-gray-200 text-gray-800",
            )}
        >
            {item.product}
        </div>
        <p className="mt-2 text-sm dark:text-gray-400 text-gray-600">
            {item.description}
        </p>

        <div className="pt-6 flex flex-col gap-2 text-sm">
            <div className="flex items-start gap-2">
                <CheckIcon className="mt-0.5 shrink-0 text-emerald-500" />
                <span className="dark:text-gray-200 text-gray-800">
                    {item.outcome}
                </span>
            </div>
            <div className="flex items-start gap-2">
                <PlusIcon className="mt-0.5 shrink-0 rotate-45 text-gray-400" />
                <span className="dark:text-gray-500 text-gray-500">
                    No extra licence for {item.instead}
                </span>
            </div>
            {item.link && (
                <Link
                    to={item.link}
                    className={clsx(
                        "mt-2 text-sm font-semibold",
                        "dark:text-refine-cyan-alt text-refine-blue",
                    )}
                >
                    Learn more →
                </Link>
            )}
        </div>
        {item.image && (
            <div className="mt-auto pt-6">
                <ScreenshotFrame src={item.image} alt={item.imageAlt ?? item.product} />
            </div>
        )}
    </div>
);

const Groups = ({ groups }: { groups: SolutionGroup[] }) => {
    let counter = 0;
    return (
        <>
            {groups.map((g, gi) => (
                <div
                    key={g.title}
                    className={clsx(sectionWidth, "not-prose")}
                >
                    <SectionHeading
                        id={gi === 0 ? "stack" : undefined}
                        title={g.title}
                        subtitle={g.subtitle}
                    />
                    <div
                        className={clsx(
                            "mt-8 landing-md:mt-12",
                            "grid grid-cols-1 landing-md:grid-cols-2",
                            "gap-4 landing-sm:gap-6",
                        )}
                    >
                        {g.items.map((item) => {
                            counter += 1;
                            return (
                                <ChallengeCard
                                    key={item.title}
                                    item={item}
                                    index={counter}
                                />
                            );
                        })}
                    </div>
                </div>
            ))}
        </>
    );
};

const Columns = ({
    title,
    columns,
}: {
    title: string;
    columns: SolutionColumn[];
}) => (
    <div className={clsx(sectionWidth, "not-prose")}>
        <SectionHeading title={title} />
        <div
            className={clsx(
                "mt-8 landing-md:mt-12",
                "grid grid-cols-1 landing-md:grid-cols-3",
                "gap-4 landing-sm:gap-6",
            )}
        >
            {columns.map((c, i) => (
                <div key={c.title} className={clsx(cardBase, "p-6 landing-sm:p-8")}>
                    <div
                        className={clsx(
                            "text-[40px] leading-none font-semibold",
                            "dark:text-gray-700 text-gray-300",
                        )}
                    >
                        {String(i + 1).padStart(2, "0")}
                    </div>
                    <h3
                        className={clsx(
                            "mt-6 text-lg font-semibold",
                            "dark:text-gray-0 text-gray-900",
                        )}
                    >
                        {c.title}
                    </h3>
                    <p className="mt-2 text-sm dark:text-gray-400 text-gray-600">
                        {c.text}
                    </p>
                </div>
            ))}
        </div>
    </div>
);

const Compare = ({
    title,
    subtitle,
    rows,
}: {
    title: string;
    subtitle: string;
    rows: SolutionCompareRow[];
}) => (
    <div className={clsx(sectionWidth, "not-prose")}>
        <SectionHeading title={title} subtitle={subtitle} />
        <div
            className={clsx(
                cardBase,
                "mt-8 landing-md:mt-12",
                "overflow-hidden",
            )}
        >
            <div
                className={clsx(
                    "grid grid-cols-[1.4fr_1fr_1fr]",
                    "px-4 landing-sm:px-8 py-4",
                    "text-xs landing-sm:text-sm font-semibold",
                    "dark:text-gray-400 text-gray-600",
                    "border-b dark:border-gray-700 border-gray-200",
                )}
            >
                <div>What you need</div>
                <div>Typical panel</div>
                <div className="dark:text-refine-cyan-alt text-refine-blue">
                    OpenPanel
                </div>
            </div>
            {rows.map((r, i) => (
                <div
                    key={r.feature}
                    className={clsx(
                        "grid grid-cols-[1.4fr_1fr_1fr] items-center gap-2",
                        "px-4 landing-sm:px-8 py-3 landing-sm:py-4",
                        "text-xs landing-sm:text-sm",
                        i < rows.length - 1 &&
                            "border-b dark:border-gray-700/60 border-gray-200/70",
                    )}
                >
                    <div className="font-medium dark:text-gray-200 text-gray-800">
                        {r.feature}
                    </div>
                    <div className="flex items-center gap-2 dark:text-gray-400 text-gray-600">
                        <PlusIcon className="shrink-0 text-orange-500" />
                        <span>{r.typical}</span>
                    </div>
                    <div className="flex items-center gap-2 dark:text-gray-0 text-gray-900">
                        <CheckIcon className="shrink-0 text-emerald-500" />
                        <span>Included</span>
                    </div>
                </div>
            ))}
            <div
                className={clsx(
                    "grid grid-cols-[1.4fr_1fr_1fr] items-center gap-2",
                    "px-4 landing-sm:px-8 py-4",
                    "text-xs landing-sm:text-sm font-semibold",
                    "dark:bg-gray-900/40 bg-white",
                )}
            >
                <div className="dark:text-gray-0 text-gray-900">
                    What you pay
                </div>
                <div className="dark:text-gray-400 text-gray-600">
                    Panel + one licence per add-on, often priced per account
                </div>
                <div className={accent}>€14.95 / server / month</div>
            </div>
        </div>
    </div>
);

const COMPARE_PAGES = [
    { label: "cPanel/WHM", to: "/cpanel-alternative/" },
    { label: "Plesk", to: "/plesk-alternative/" },
    { label: "DirectAdmin", to: "/directadmin-alternative/" },
    { label: "CyberPanel", to: "/cyberpanel-alternative/" },
    { label: "Enhance", to: "/enhance-alternative/" },
];

const COMPARE_INTRO =
    "OpenPanel isolates every user in their own context and runs each of their services in a separate container, and includes the WAF, malware scanning, backups, WordPress manager, resellers and branding in the base licence, for a flat €14.95 per server with unlimited accounts. Many other panels sell isolation, security or backups as separate add-ons, or price by the number of accounts.";

// Appended to the FAQ of every solutions page.
const COMPARE_FAQ: SolutionFaq = {
    question: "How is OpenPanel different from other control panels?",
    answer: `${COMPARE_INTRO} Side-by-side comparisons: ${COMPARE_PAGES.map(
        (c) => `OpenPanel vs ${c.label} (https://openpanel.com${c.to})`,
    ).join(", ")}.`,
    answerNode: (
        <>
            {COMPARE_INTRO} See the side-by-side comparisons:{" "}
            {COMPARE_PAGES.map((c, i) => (
                <React.Fragment key={c.to}>
                    <Link
                        to={c.to}
                        className={clsx(
                            "font-semibold underline underline-offset-2",
                            "dark:text-refine-cyan-alt text-refine-blue",
                        )}
                    >
                        OpenPanel vs {c.label}
                    </Link>
                    {i < COMPARE_PAGES.length - 2
                        ? ", "
                        : i === COMPARE_PAGES.length - 2
                          ? " and "
                          : "."}
                </React.Fragment>
            ))}
        </>
    ),
};

const Faq = ({ title, faq }: { title: string; faq: SolutionFaq[] }) => {
    const [open, setOpen] = useState<number | null>(0);
    return (
        <div
            className={clsx(
                "not-prose",
                "px-4 landing-sm:px-10 landing-lg:px-0",
                "w-full landing-lg:max-w-[792px] mx-auto",
            )}
        >
            <h2
                className={clsx(
                    "font-semibold",
                    "text-2xl landing-sm:text-[32px] landing-sm:leading-[40px]",
                    "dark:text-gray-0 text-gray-900",
                )}
            >
                {title}
            </h2>
            <div className="mt-8 flex flex-col gap-3">
                {faq.map((f, i) => {
                    const isOpen = open === i;
                    return (
                        <div key={f.question} className={clsx(cardBase)}>
                            <button
                                type="button"
                                onClick={() => setOpen(isOpen ? null : i)}
                                aria-expanded={isOpen}
                                className={clsx(
                                    "w-full text-left",
                                    "flex items-center justify-between gap-4",
                                    "px-5 landing-sm:px-6 py-4",
                                    "text-sm landing-sm:text-base font-semibold",
                                    "dark:text-gray-0 text-gray-900",
                                    "bg-transparent border-0 cursor-pointer",
                                )}
                            >
                                <span>{f.question}</span>
                                <span
                                    className={clsx(
                                        "shrink-0 text-xl leading-none transition-transform",
                                        "dark:text-gray-400 text-gray-500",
                                        isOpen && "rotate-45",
                                    )}
                                >
                                    +
                                </span>
                            </button>
                            {isOpen && (
                                <p
                                    className={clsx(
                                        "px-5 landing-sm:px-6 pb-5 -mt-1",
                                        "text-sm",
                                        "dark:text-gray-400 text-gray-600",
                                    )}
                                >
                                    {f.answerNode ?? f.answer}
                                </p>
                            )}
                        </div>
                    );
                })}
            </div>
        </div>
    );
};

const ClosingCta = ({
    title,
    text,
    eventPrefix,
}: {
    title: string;
    text: string;
    eventPrefix: string;
}) => (
    <div className={clsx(sectionWidth, "not-prose", "landing-lg:max-w-[1000px]")}>
        <div
            className={clsx(
                "rounded-3xl",
                "px-6 py-10 landing-sm:px-12 landing-sm:py-14",
                "text-center",
                "dark:bg-gray-800 bg-gray-50",
                "dark:bg-enterprise-cta-dark bg-enterprise-cta-light",
            )}
        >
            <h2
                className={clsx(
                    "font-semibold",
                    "text-2xl landing-sm:text-[36px] landing-sm:leading-[44px]",
                    "dark:text-gray-0 text-gray-900",
                )}
            >
                {title}
            </h2>
            <p className="mt-4 mx-auto max-w-[560px] dark:text-gray-400 text-gray-600">
                {text}
            </p>
            <div
                className={clsx(
                    "mt-8",
                    "flex flex-col landing-sm:flex-row",
                    "items-center justify-center",
                    "gap-4 landing-sm:gap-6",
                )}
            >
                <EnterpriseGetInTouchButton
                    label="Start free 30-day trial"
                    href="/trial"
                    eventName={`trial_click_cta_${eventPrefix}`}
                />
                <Link
                    to="/demo"
                    className={clsx(
                        "text-sm font-semibold",
                        "dark:text-gray-0 text-gray-900",
                        "underline underline-offset-4",
                    )}
                >
                    Try the live demo →
                </Link>
            </div>
            <Link
                to="mailto:info@openpanel.com"
                className="mt-6 inline-block text-sm dark:text-gray-500 text-gray-500"
            >
                Prefer to talk first? info@openpanel.com
            </Link>
        </div>
    </div>
);

export const SolutionPage = (props: SolutionPageProps) => {
    const p = { ...props, faq: [...props.faq, COMPARE_FAQ] };
    return (
    <>
        <Head>
            <html data-active-page="index" />
            <title>{p.metaTitle}</title>
            <meta property="og:title" content={p.metaTitle} />
        </Head>
        <FaqSchema faq={p.faq} />
        <CommonLayout description={p.metaDescription}>
            <div>
                <CommonHeader />
                <div
                    className={clsx(
                        "flex flex-col",
                        "gap-16 landing-sm:gap-20 landing-md:gap-28 landing-lg:gap-32",
                        "pb-12 landing-sm:pb-16 landing-md:pb-20 landing-lg:pb-32",
                        "mx-auto",
                    )}
                >
                    <Hero {...p} />
                    <Stats stats={p.stats} />
                    <Groups groups={p.groups} />
                    <Columns title={p.columnsTitle} columns={p.columns} />
                    <Compare
                        title={p.compareTitle}
                        subtitle={p.compareSubtitle}
                        rows={p.compareRows}
                    />
                    {p.extra}
                    <Faq title={p.faqTitle} faq={p.faq} />
                    <ClosingCta
                        title={p.ctaTitle}
                        text={p.ctaText}
                        eventPrefix={p.eventPrefix}
                    />
                </div>
                <LandingFooter />
            </div>
        </CommonLayout>
    </>
    );
};

// Row of billing-system integrations, used on the hosting providers page.
export const SolutionBillingStrip = () => (
    <div className={clsx(sectionWidth, "not-prose")}>
        <SectionHeading
            title="Plugs into the billing system you already run"
            subtitle="Free provisioning modules create, suspend, upgrade and terminate OpenPanel accounts straight from your billing panel."
        />
        <div
            className={clsx(
                "mt-8 landing-md:mt-12",
                "grid grid-cols-2 landing-sm:grid-cols-5",
                "gap-4",
            )}
        >
            {[
                { label: "WHMCS", to: "/billing/whmcs-module/" },
                { label: "Blesta", to: "/billing/blesta-module/" },
                { label: "FOSSBilling", to: "/billing/fossbilling-module/" },
                { label: "ClientExec", to: "/billing/clientexec-module/" },
                { label: "WiseCP", to: "/billing/wisecp-module/" },
            ].map((item) => (
                <Link
                    key={item.to}
                    to={item.to}
                    className={clsx(
                        cardBase,
                        "flex items-center justify-center text-center",
                        "p-4 landing-sm:p-6",
                        "text-sm landing-sm:text-base font-semibold",
                        "text-gray-900 dark:text-gray-0",
                        "transition-[filter] duration-150 ease-in-out",
                        "hover:brightness-110 hover:!no-underline",
                    )}
                >
                    {item.label}
                </Link>
            ))}
        </div>
    </div>
);
