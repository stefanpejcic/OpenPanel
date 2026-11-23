import Head from "@docusaurus/Head";
import Link from "@docusaurus/Link";
import clsx from "clsx";
import React from "react";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";
import { HeaderDiscordIcon } from "@site/src/refine-theme/icons/header-discord";
import { FooterGithubIcon } from "@site/src/refine-theme/icons/footer-github";
import { team } from "../../assets/team";
import { inProgress, plannedFeatures } from "../../assets/roadmap";
import { WizardsIcon } from "@site/src/components/landing/icons";

const simpleIconProps = {
    width: 24,
    height: 24,
    viewBox: "0 0 24 24",
    fill: "none",
    stroke: "currentColor",
    strokeWidth: 1.5,
    strokeLinecap: "round" as const,
    strokeLinejoin: "round" as const,
};

const FlagIcon = () => (
    <svg {...simpleIconProps}>
        <line x1="6" y1="3" x2="6" y2="21" />
        <path d="M6 4h11l-3 4 3 4H6" />
    </svg>
);

const ContainerIcon = () => (
    <svg {...simpleIconProps}>
        <path d="M12 3l8 4.5v9L12 21l-8-4.5v-9L12 3Z" />
        <path d="M12 12v9" />
        <path d="M4 7.5L12 12l8-4.5" />
    </svg>
);

const ShieldIcon = () => (
    <svg {...simpleIconProps}>
        <path d="M12 3l7 3v6c0 4.5-3 7.5-7 9-4-1.5-7-4.5-7-9V6l7-3Z" />
    </svg>
);

const ControlIcon = () => (
    <svg {...simpleIconProps}>
        <line x1="5" y1="4" x2="5" y2="20" />
        <circle cx="5" cy="9" r="2" />
        <line x1="12" y1="4" x2="12" y2="20" />
        <circle cx="12" cy="15" r="2" />
        <line x1="19" y1="4" x2="19" y2="20" />
        <circle cx="19" cy="7" r="2" />
    </svg>
);

const ServerIcon = () => (
    <svg {...simpleIconProps}>
        <rect x="4" y="4" width="16" height="7" rx="1.5" />
        <rect x="4" y="13" width="16" height="7" rx="1.5" />
        <circle cx="8" cy="7.5" r="1" fill="currentColor" stroke="none" />
        <circle cx="8" cy="16.5" r="1" fill="currentColor" stroke="none" />
    </svg>
);

const stats = [
    { value: "100%", label: "Self-funded & independent" },
    { value: "€14.95", label: "Fixed pricing, from day one" },
    { value: "$0", label: "Per-user or per-site fees" },
    { value: "Free", label: "Forever free edition" },
];

const timeline = [
    {
        year: "2023",
        achievements: [
            {
                title: "Beta edition",
                description:
                    "Started as a side project, paid for out of our own pocket, because we were unhappy with the panels we used.",
            },
        ],
    },
    {
        year: "2024",
        achievements: [
            {
                title: "OpenPanel 0.3",
                link: "/blog/openpanel-public-beta-released",
                description:
                    "First public beta. It already ran every user in Docker.",
            },
            {
                title: "openpanel.com",
                description:
                    "Bought openpanel.com and moved the project over from openpanel.co.",
            },
        ],
    },
    {
        year: "2025",
        achievements: [
            {
                title: "OpenPanel 1.0",
                description:
                    "Launched OpenPanel Enterprise, with premium features and priority support for hosting providers.",
            },
        ],
    },
    {
        year: "2026",
        achievements: [
            {
                title: "OpenPanel LLC",
                description:
                    "Formalized the company, incorporating OpenPanel as an LLC.",
            },
            {
                title: "100,000+ installations",
                description: "Crossed 100,000 active installations worldwide.",
            },
            {
                title: "OpenPanel 2.0",
                link: "/blog/podman-switch-openpanel-2",
                description:
                    "Core rewritten in Go. Podman replaced Docker, so user containers run rootless.",
            },
            {
                title: "Multi-Server Clustering",
                link: "/blog/multi-server-openpanel",
                description:
                    "Manage several OpenPanel servers from one OpenAdmin, with settings synced between them and accounts moved from one server to another.",
            },
        ],
    },
];

const story = [
    {
        icon: <WizardsIcon />,
        title: "The idea",
        description:
            'At UNLIMITED.RS we ran shared hosting and managed VPS on commercial panels. The license price went up every year and the panels did not get any more flexible. During one bad migration a colleague said "we could build something better ourselves". Our co-founder and CEO, Stefan Pejčić, took that seriously.',
    },
    {
        icon: <FlagIcon />,
        title: "Early days",
        description:
            "Work started in 2023 on a simple LAMP panel for 10 to 500 users per server. First we tried a Docker container per website. One WordPress site needed three containers (web, PHP, database), and that was too much overhead. So we moved to one container per user. It was less flexible but much more stable. Later we split that again into one container per service of each user, so a user's web server, database and mail are separated without the cost we had with the per-website setup.",
    },
    {
        icon: <ContainerIcon />,
        title: "The 2.0 rewrite",
        description:
            "All 1.x versions ran on Docker. For 2.0 we rewrote the container layer for Podman. There is no Docker daemon anymore, containers run rootless, and every feature from 1.x still works.",
    },
];

const beliefs = [
    {
        icon: <ContainerIcon />,
        title: "Isolation is in every edition",
        description:
            "Each user gets their own MySQL, their own PHP version and their own resource limits, also in the free Community Edition. One busy site shouldn't slow down everyone else on the server.",
    },
    {
        icon: <ShieldIcon />,
        title: "Security is on after install",
        description:
            "The WAF, malware scanner, two-factor login and IP blocking are part of the panel. You don't have to install them separately after setup.",
    },
    {
        icon: <ControlIcon />,
        title: "You decide what users see",
        description:
            "Pick the web server, set the branding, hide features you don't sell, and set your own prices. We don't take a cut per account.",
    },
    {
        icon: <ServerIcon />,
        title: "It runs on your server",
        description:
            "OpenPanel is installed on your server and your data stays there. Nothing goes through our servers, and the panel doesn't depend on us being online.",
    },
];

const getInitials = (name: string) =>
    name
        .split(" ")
        .map((part) => part[0])
        .join("")
        .slice(0, 2)
        .toUpperCase();


const card = "rounded-2xl border border-gray-300 dark:border-gray-700 p-5 sm:p-6";
const h2 = "text-2xl sm:text-3xl font-bold text-gray-900 dark:text-gray-0 m-0";
const lead = "text-base text-gray-600 dark:text-gray-400 m-0";
const primaryBtn =
    "inline-flex items-center gap-2 rounded-full px-5 py-2.5 text-sm font-semibold bg-refine-blue dark:bg-refine-cyan-alt text-gray-0 dark:text-gray-900 hover:brightness-110 no-underline";
const secondaryBtn =
    "inline-flex items-center gap-2 rounded-full px-5 py-2.5 text-sm font-semibold border border-gray-300 dark:border-gray-700 text-gray-900 dark:text-gray-100 hover:bg-gray-100 dark:hover:bg-gray-800 no-underline";
const link = "text-refine-react-light-link dark:text-refine-react-dark-link hover:underline";

const GITHUB = "https://github.com/stefanpejcic/openpanel";
const DISCORD = "https://discord.com/invite/7bNY8fANqF";

const About: React.FC = () => {
    return (
        <CommonLayout
            description="OpenPanel is made by a small team that also runs a hosting company. Who we are, how the project started and what we're working on."
            image="/img/og/about.png"
        >
            <Head>
                <title>About | OpenPanel Project</title>
                <meta property="og:title" content="About | OpenPanel Project" />
                <meta property="og:image:width" content="1200" />
                <meta property="og:image:height" content="630" />
                <meta property="og:image:type" content="image/png" />
                <html data-page="about" data-customized="true" />
            </Head>
            <div className="refine-prose">
                <CommonHeader hasSticky={true} />

                <div className="not-prose flex flex-col gap-16 pt-8 lg:pt-16 pb-32 max-w-[928px] landing-lg:max-w-[1232px] w-full mx-auto px-4">
                    <section className="flex flex-col gap-5">
                        <h1 className="text-3xl sm:text-5xl font-bold text-gray-900 dark:text-gray-0 m-0 max-w-[860px]">
                            We run a hosting company, and we built the panel we wanted to use
                        </h1>
                        <p className="text-base sm:text-lg text-gray-600 dark:text-gray-400 m-0 max-w-[720px]">
                            OpenPanel started inside UNLIMITED.RS, a hosting company in Serbia. The price per server stays the same whether
                            you host 10 accounts or 1,000.
                        </p>
                        <div className="flex flex-wrap gap-3">
                            <Link to={GITHUB} className={primaryBtn}>
                                <FooterGithubIcon className="w-4 h-4" /> Star on GitHub
                            </Link>
                            <Link to={DISCORD} className={secondaryBtn}>
                                <HeaderDiscordIcon className="w-4 h-4" /> Join Discord
                            </Link>
                        </div>
                        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 mt-2">
                            {stats.map((s) => (
                                <div key={s.label} className="rounded-xl border border-gray-200 dark:border-gray-700 p-4">
                                    <div className="text-2xl font-bold text-gray-900 dark:text-gray-0">{s.value}</div>
                                    <div className="text-sm text-gray-600 dark:text-gray-400">{s.label}</div>
                                </div>
                            ))}
                        </div>
                    </section>

                    <section className="flex flex-col gap-6">
                        <div className="flex flex-col gap-2">
                            <h2 className={h2}>Where we&apos;ve been</h2>
                            <p className={lead}>The main dates, from the first beta to today.</p>
                        </div>
                        <ol className="m-0 p-0 list-none flex flex-col">
                            {timeline.map((t, i) => (
                                <li key={t.year} className="grid grid-cols-[64px_1fr] sm:grid-cols-[96px_1fr] gap-4">
                                    <div className="flex flex-col items-start">
                                        <span className="text-lg sm:text-xl font-bold text-refine-blue dark:text-refine-cyan-alt">{t.year}</span>
                                    </div>
                                    <div
                                        className={clsx(
                                            "border-l border-gray-300 dark:border-gray-700 pl-5 flex flex-col gap-4",
                                            i < timeline.length - 1 ? "pb-8" : "pb-1",
                                        )}
                                    >
                                        {t.achievements.map((a) => (
                                            <div key={a.title} className="relative flex flex-col gap-1">
                                                <span className="absolute -left-[25px] top-1.5 w-2.5 h-2.5 rounded-full bg-refine-blue dark:bg-refine-cyan-alt" />
                                                {a.link ? (
                                                    <Link to={a.link} className="text-sm font-semibold text-gray-900 dark:text-gray-100 hover:underline no-underline">
                                                        {a.title} →
                                                    </Link>
                                                ) : (
                                                    <span className="text-sm font-semibold text-gray-900 dark:text-gray-100">{a.title}</span>
                                                )}
                                                <span className="text-sm text-gray-600 dark:text-gray-400">{a.description}</span>
                                            </div>
                                        ))}
                                    </div>
                                </li>
                            ))}
                        </ol>
                    </section>

                    {(inProgress.length > 0 || plannedFeatures.length > 0) && (
                        <section className="flex flex-col gap-6">
                            <div className="flex flex-col gap-2">
                                <h2 className={h2}>Where we&apos;re heading</h2>
                                <p className={lead}>What we&apos;re working on now and what comes after.</p>
                            </div>
                            <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
                                {inProgress.map((f) => (
                                    <div key={f.title} className={clsx(card, "flex flex-col gap-1 !p-4")}>
                                        <span className="text-xs font-semibold uppercase tracking-wide text-refine-blue dark:text-refine-cyan-alt">In progress</span>
                                        <span className="text-sm font-semibold text-gray-900 dark:text-gray-100">{f.title}</span>
                                        <span className="text-sm text-gray-600 dark:text-gray-400">{f.description}</span>
                                    </div>
                                ))}
                                {plannedFeatures.map((f) => (
                                    <div key={f.title} className={clsx(card, "flex flex-col gap-1 !p-4")}>
                                        <span className="text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-gray-400">Planned</span>
                                        <span className="text-sm font-semibold text-gray-900 dark:text-gray-100">{f.title}</span>
                                    </div>
                                ))}
                            </div>
                            <Link to="/roadmap/" className={clsx("self-start text-sm font-semibold", link)}>
                                See the full roadmap →
                            </Link>
                        </section>
                    )}

                    <section className="flex flex-col gap-6">
                        <div className="flex flex-col gap-2">
                            <h2 className={h2}>Our story</h2>
                            <p className={lead}>How the project started and why it runs on containers.</p>
                        </div>
                        <div className="grid gap-4 lg:grid-cols-3">
                            {story.map((s) => (
                                <div key={s.title} className={clsx(card, "flex flex-col gap-3")}>
                                    <div className="text-refine-blue dark:text-refine-cyan-alt">{s.icon}</div>
                                    <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-0 m-0">{s.title}</h3>
                                    <p className="text-sm text-gray-600 dark:text-gray-400 m-0 leading-relaxed">{s.description}</p>
                                </div>
                            ))}
                        </div>
                    </section>

                    <section className="flex flex-col gap-6">
                        <div className="flex flex-col gap-2">
                            <h2 className={h2}>What we believe</h2>
                            <p className={lead}>A few rules we try to follow.</p>
                        </div>
                        <div className="grid gap-4 sm:grid-cols-2">
                            {beliefs.map((b) => (
                                <div key={b.title} className={clsx(card, "flex gap-4")}>
                                    <div className="shrink-0 text-refine-blue dark:text-refine-cyan-alt">{b.icon}</div>
                                    <div className="flex flex-col gap-1.5">
                                        <h3 className="text-base font-semibold text-gray-900 dark:text-gray-0 m-0">{b.title}</h3>
                                        <p className="text-sm text-gray-600 dark:text-gray-400 m-0">{b.description}</p>
                                    </div>
                                </div>
                            ))}
                        </div>
                    </section>

                    <section className={clsx(card, "flex flex-col gap-4 sm:!p-8")}>
                        <h2 className="text-xl sm:text-2xl font-bold text-gray-900 dark:text-gray-0 m-0">Independent by choice</h2>
                        <p className="text-base text-gray-600 dark:text-gray-400 m-0 max-w-[760px]">
                            We&apos;ve had offers from competitors and from investors. We said no, because every offer came with
                            conditions we didn&apos;t want for the product. We&apos;re still self-funded, and
                            the promise stays: <b className="text-gray-900 dark:text-gray-0">we will never charge per user or per website</b>.
                        </p>
                        <div className="flex flex-wrap gap-3">
                            <Link to="/community/" className={primaryBtn}>
                                Community Edition
                            </Link>
                            <Link to="/enterprise" className={secondaryBtn}>
                                Enterprise Edition
                            </Link>
                        </div>
                    </section>

                    <section className="flex flex-col gap-6">
                        <div className="flex flex-col gap-2">
                            <h2 className={h2}>Who&apos;s behind it</h2>
                            <p className={lead}>OpenPanel is our full-time job.</p>
                        </div>
                        <div className="grid gap-4 grid-cols-2 lg:grid-cols-4">
                            {team.map(({ name, avatar, role1, role2, linkedin }) => (
                                <Link
                                    key={name}
                                    to={linkedin}
                                    className={clsx(card, "flex flex-col items-center text-center gap-3 no-underline hover:border-gray-400 dark:hover:border-gray-500")}
                                >
                                    {avatar ? (
                                        <img src={avatar} alt={name} loading="lazy" className="w-20 h-20 rounded-full object-cover" />
                                    ) : (
                                        <span className="w-20 h-20 rounded-full bg-gray-100 dark:bg-gray-800 flex items-center justify-center text-xl font-bold text-gray-700 dark:text-gray-200">
                                            {getInitials(name)}
                                        </span>
                                    )}
                                    <div className="flex flex-col">
                                        <span className="text-sm font-semibold text-gray-900 dark:text-gray-100">{name}</span>
                                        <span className="text-xs text-gray-600 dark:text-gray-400">{role1}</span>
                                        {role2 && <span className="text-xs text-gray-600 dark:text-gray-400">{role2}</span>}
                                    </div>
                                </Link>
                            ))}
                        </div>
                    </section>

                    <section className={clsx(card, "flex flex-col sm:flex-row gap-4 sm:items-center sm:justify-between sm:!p-8")}>
                        <div className="flex flex-col gap-1">
                            <h2 className="text-xl font-bold text-gray-900 dark:text-gray-0 m-0">Want to help?</h2>
                            <p className="text-sm text-gray-600 dark:text-gray-400 m-0">
                                Join our community, contribute on GitHub, or just say hello at{" "}
                                <Link to="mailto:info@openpanel.com" className={link}>
                                    info@openpanel.com
                                </Link>
                                .
                            </p>
                        </div>
                        <div className="flex flex-wrap gap-3">
                            <Link to={DISCORD} className={primaryBtn}>
                                <HeaderDiscordIcon className="w-4 h-4" /> Join Discord
                            </Link>
                            <Link to={GITHUB} className={secondaryBtn}>
                                <FooterGithubIcon className="w-4 h-4" /> View on GitHub
                            </Link>
                        </div>
                    </section>
                </div>
                <BlogFooter />
            </div>
        </CommonLayout>
    );
};

export default About;
