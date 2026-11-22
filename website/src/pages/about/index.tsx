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
                    "Started as a bootstrapped side project to fix the hosting panels we were stuck using ourselves.",
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
                    "First public beta release: Docker-native from day one, not bolted on later.",
            },
            {
                title: "openpanel.com",
                description:
                    "Acquired openpanel.com and moved over from openpanel.co, giving the project a proper permanent home.",
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
                    "Rewrote the core in Go to run on Podman, replacing Docker as the container engine, for rootless isolation by default.",
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
            'OpenPanel was born out of frustration with the premium hosting panels we relied on at UNLIMITED.RS to run shared hosting and managed VPS for our own clients. Prices kept climbing every year, but the flexibility never did. During one particularly rough migration, a colleague joked, "Even we could build something better." The comment stuck - and sent our co-founder and CEO, Stefan Pejčić, down the path of actually doing it.',
    },
    {
        icon: <FlagIcon />,
        title: "Early days",
        description:
            "Work started in 2023 on a basic LAMP-stack panel, built to comfortably run 10 to 500 users on a single server. The first instinct was to isolate every website in its own Docker container - until a single WordPress install alone needed three containers, one per service, and the overhead stopped making sense. We pivoted to one container per user instead, trading a bit of flexibility for a far more stable, predictable environment. Later, we pivoted again - to one container per user service - giving every user real isolation between their web, database, and mail without repeating the overhead that sank the per-website approach.",
    },
    {
        icon: <ContainerIcon />,
        title: "The 2.0 rewrite",
        description:
            "OpenPanel ran on Docker from that first pivot all the way through 1.x. For 2.0, we rewrote the container layer on Podman - dropping the Docker daemon in favor of rootless, daemonless containers, with no functionality lost along the way.",
    },
];

const beliefs = [
    {
        icon: <ContainerIcon />,
        title: "Isolation is a right, not a premium feature",
        description:
            "Every user on your server deserves their own environment - their own MySQL, their own PHP version, their own resources - without interference from anyone else. That's not enterprise-tier, that's just how hosting should work.",
    },
    {
        icon: <ShieldIcon />,
        title: "Security should be built in, not bolted on",
        description:
            "WAF protection, malware scanning, two-factor authentication, IP controls - these shouldn't be things you configure after the fact. They're there from the moment you install.",
    },
    {
        icon: <ControlIcon />,
        title: "Control panels should give you control",
        description:
            "Over your stack, your branding, your pricing, your data. You're running the server - the panel should work for you, not the other way around.",
    },
    {
        icon: <ServerIcon />,
        title: "Self-hosted, not SaaS",
        description:
            "OpenPanel runs entirely on your own server. Your data, your infrastructure, your rules - nothing routed through us, nothing held hostage if we disappear tomorrow.",
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
            description="OpenPanel is built by hosting operators, for hosting operators - Podman-native isolation, built-in security, and no per-seat pricing."
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
                            Built by hosting operators who got tired of paying more for less
                        </h1>
                        <p className="text-base sm:text-lg text-gray-600 dark:text-gray-400 m-0 max-w-[720px]">
                            OpenPanel is a Podman-native control panel built by hosting operators, for hosting operators: real user
                            isolation, security built in, and pricing that doesn&apos;t punish you for growing.
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
                            <p className={lead}>From a bootstrapped side project to 100,000+ active installations, the short version.</p>
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
                                <p className={lead}>What we&apos;re building right now, and what&apos;s up next.</p>
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
                            <p className={lead}>Containers had changed everything about infrastructure, but hosting control panels were still stuck in the past.</p>
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
                            <p className={lead}>The principles that shape every decision we make about OpenPanel.</p>
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
                            Over the years we turned down offers from competitors and venture capitalists. Every one came with
                            conditions that would&apos;ve compromised the product, so we stayed self-funded instead. One promise we
                            stand by: <b className="text-gray-900 dark:text-gray-0">we will never charge per user or per website</b>.
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
                            <p className={lead}>A small, independent team building OpenPanel full-time.</p>
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
                            <h2 className="text-xl font-bold text-gray-900 dark:text-gray-0 m-0">Come build the future of hosting with us</h2>
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
