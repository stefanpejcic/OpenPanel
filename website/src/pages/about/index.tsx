import Head from "@docusaurus/Head";
import Link from "@docusaurus/Link";
import clsx from "clsx";
import React from "react";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";
import { HeaderDiscordIcon } from "@site/src/refine-theme/icons/header-discord";
import { FooterGithubIcon } from "@site/src/refine-theme/icons/footer-github";
import { MailIcon } from "@site/src/refine-theme/icons/mail";
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
                title: "OpenPanel is born",
                description:
                    "Started as a bootstrapped side project to fix the hosting panels we were stuck using ourselves.",
            },
        ],
    },
    {
        year: "2024",
        achievements: [
            {
                title: "Public beta",
                description:
                    "First public beta release - Docker-native from day one, not bolted on later.",
            },
        ],
    },
    {
        year: "2025",
        achievements: [
            {
                title: "openpanel.com",
                description: "Acquired the openpanel.com domain.",
            },
            {
                title: "Enterprise Edition",
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

const gradientAccentA = clsx(
    "text-transparent bg-clip-text bg-gradient-to-r text-gradient-to-r",
    "from-[#0FBDBD] to-[#26D97F]",
);

const gradientAccentB = clsx(
    "text-transparent bg-clip-text bg-gradient-to-r text-gradient-to-r",
    "from-[#FF9933] to-[#FF4C4D]",
);

const sectionHeading = clsx(
    "text-2xl landing-sm:text-[32px] landing-sm:leading-[40px]",
    "tracking-tight",
    "p-0",
    "dark:text-gray-0 text-gray-900",
);

const sectionAccent = clsx(
    "font-semibold",
    "dark:text-refine-cyan-alt dark:drop-shadow-[0_0_30px_rgba(71,235,235,0.25)]",
    "text-refine-blue drop-shadow-[0_0_30px_rgba(0,128,255,0.3)]",
);

const cardShell = clsx(
    "not-prose",
    "p-4 landing-sm:p-10",
    "flex",
    "flex-col landing-sm:flex-row",
    "items-start",
    "gap-6",
    "dark:bg-landing-noise",
    "dark:bg-gray-800 bg-gray-50",
    "rounded-2xl landing-sm:rounded-3xl",
);

const About: React.FC = () => {
    return (
        <CommonLayout description="OpenPanel is built by hosting operators, for hosting operators - Docker-native isolation, built-in security, and no per-seat pricing.">
            <Head title="About | OpenPanel Project">
                <html data-page="about" data-customized="true" />
            </Head>
            <div className="refine-prose">
                <CommonHeader hasSticky={true} />

                <div
                    className={clsx(
                        "flex flex-col",
                        "gap-16 landing-sm:gap-20 landing-md:gap-28",
                        "w-full max-w-[592px] landing-sm:max-w-[656px] landing-md:max-w-[896px] landing-lg:max-w-[1200px]",
                        "px-2 landing-sm:px-0",
                        "pt-8 landing-sm:pt-12 landing-lg:pt-20",
                        "pb-16 landing-sm:pb-24 landing-md:pb-32",
                        "mx-auto",
                        "not-prose",
                    )}
                >
                    {/* Hero */}
                    <div className={clsx("px-4 landing-md:px-10")}>
                        <h1
                            className={clsx(
                                "text-3xl landing-sm:text-[40px] landing-sm:leading-[48px]",
                                "font-semibold",
                                "tracking-tight",
                                "text-center",
                                "max-w-3xl mx-auto",
                                "p-0",
                                "dark:text-gray-0 text-gray-900",
                            )}
                        >
                            Built by{" "}
                            <span className={gradientAccentA}>hosting operators</span>{" "}
                            who got tired of{" "}
                            <span className={gradientAccentB}>
                                paying more for less
                            </span>
                            .
                        </h1>
                        <p
                            className={clsx(
                                "mt-4 landing-sm:mt-6",
                                "max-w-lg mx-auto",
                                "text-base text-center",
                                "dark:text-gray-400 text-gray-600",
                            )}
                        >
                            OpenPanel is a Docker-native control panel built by hosting
                            operators, for hosting operators - real user isolation,
                            security built in, and pricing that doesn&apos;t punish you
                            for growing.
                        </p>
                        <div
                            className={clsx(
                                "mt-8",
                                "flex flex-col landing-xs:flex-row",
                                "items-center justify-center",
                                "gap-4",
                            )}
                        >
                            <Link
                                to="https://github.com/stefanpejcic/openpanel"
                                target="_blank"
                                rel="noopener noreferrer"
                                className={clsx(
                                    "w-full landing-xs:w-auto",
                                    "!text-gray-0 dark:!text-gray-900",
                                    "bg-refine-blue dark:bg-refine-cyan-alt",
                                    "transition-[filter]",
                                    "duration-150",
                                    "ease-in-out",
                                    "hover:brightness-110",
                                    "hover:!no-underline",
                                    "rounded-3xl",
                                    "py-3 px-8",
                                    "flex items-center justify-center gap-2",
                                    "text-base font-semibold",
                                )}
                            >
                                <FooterGithubIcon width={20} height={20} />
                                Star on GitHub
                            </Link>
                            <Link
                                to="https://discord.com/invite/7bNY8fANqF"
                                target="_blank"
                                rel="noopener noreferrer"
                                className={clsx(
                                    "w-full landing-xs:w-auto",
                                    "dark:text-refine-cyan-alt text-refine-blue",
                                    "dark:bg-refine-cyan-alt/10 bg-refine-blue/10",
                                    "transition-[filter]",
                                    "duration-150",
                                    "ease-in-out",
                                    "hover:brightness-110",
                                    "hover:!no-underline",
                                    "rounded-3xl",
                                    "py-3 px-8",
                                    "flex items-center justify-center gap-2",
                                    "text-base font-semibold",
                                )}
                            >
                                <HeaderDiscordIcon width={20} height={20} />
                                Join Discord
                            </Link>
                        </div>
                    </div>

                    {/* Stats */}
                    <div className={clsx("px-4 landing-md:px-10")}>
                        <div
                            className={clsx(
                                "grid",
                                "grid-cols-2 landing-sm:grid-cols-4",
                                "gap-4",
                            )}
                        >
                            {stats.map((stat, index) => (
                                <div
                                    key={index}
                                    className={clsx(
                                        "not-prose",
                                        "flex flex-col items-center text-center",
                                        "gap-2",
                                        "p-6",
                                        "dark:bg-landing-noise",
                                        "dark:bg-gray-800 bg-gray-50",
                                        "rounded-2xl",
                                    )}
                                >
                                    <div
                                        className={clsx(
                                            "text-2xl landing-sm:text-3xl",
                                            "font-semibold",
                                            "dark:text-refine-cyan-alt text-refine-blue",
                                        )}
                                    >
                                        {stat.value}
                                    </div>
                                    <div
                                        className={clsx(
                                            "text-sm",
                                            "dark:text-gray-400 text-gray-600",
                                        )}
                                    >
                                        {stat.label}
                                    </div>
                                </div>
                            ))}
                        </div>
                    </div>

                    {/* Timeline */}
                    <div className={clsx("w-full")}>
                        <div className={clsx("px-4 landing-md:px-10")}>
                            <h2 className={sectionHeading}>
                                Where we&apos;ve <span className={sectionAccent}>been</span>.
                            </h2>
                            <p
                                className={clsx(
                                    "mt-4 landing-sm:mt-6",
                                    "max-w-md",
                                    "text-base",
                                    "dark:text-gray-400 text-gray-600",
                                )}
                            >
                                From a bootstrapped side project to 100,000+ active
                                installations - the short version.
                            </p>
                        </div>

                        <div
                            className={clsx(
                                "not-prose",
                                "mt-8 landing-sm:mt-12 landing-lg:mt-20",
                                "px-4 landing-md:px-10",
                            )}
                        >
                            <div
                                className={clsx(
                                    "flex flex-col landing-lg:flex-row",
                                    "gap-6 landing-lg:gap-6",
                                )}
                            >
                                {timeline.map((item, index) => (
                                    <div
                                        key={index}
                                        className={clsx(
                                            "flex landing-lg:flex-col",
                                            "gap-4",
                                            "landing-lg:flex-1",
                                        )}
                                    >
                                        <div
                                            className={clsx(
                                                "flex flex-col items-center",
                                                "flex-shrink-0",
                                            )}
                                        >
                                            <div
                                                className={clsx(
                                                    "w-3 h-3",
                                                    "rounded-full",
                                                    "dark:bg-refine-cyan-alt bg-refine-blue",
                                                    "flex-shrink-0",
                                                )}
                                            />
                                            <div
                                                className={clsx(
                                                    "w-px flex-1",
                                                    "landing-lg:hidden",
                                                    "dark:bg-gray-800 bg-gray-100",
                                                )}
                                            />
                                            <div
                                                className={clsx(
                                                    "hidden landing-lg:block",
                                                    "h-px w-full",
                                                    "mt-[5px]",
                                                    "dark:bg-gray-800 bg-gray-100",
                                                )}
                                            />
                                        </div>
                                        <div className={clsx("pb-6 landing-lg:pb-0")}>
                                            <div
                                                className={clsx(
                                                    "text-sm",
                                                    "font-semibold",
                                                    "dark:text-refine-cyan-alt text-refine-blue",
                                                )}
                                            >
                                                {item.year}
                                            </div>
                                            {item.achievements.map(
                                                (achievement, achievementIndex) => (
                                                    <div
                                                        key={achievementIndex}
                                                        className={clsx(
                                                            achievementIndex > 0 &&
                                                                "mt-3",
                                                        )}
                                                    >
                                                        <div
                                                            className={clsx(
                                                                "mt-1",
                                                                "text-base",
                                                                "font-semibold",
                                                                "dark:text-gray-0 text-gray-900",
                                                            )}
                                                        >
                                                            {achievement.title}
                                                        </div>
                                                        <div
                                                            className={clsx(
                                                                "mt-1",
                                                                "text-sm",
                                                                "dark:text-gray-400 text-gray-600",
                                                            )}
                                                        >
                                                            {achievement.description}
                                                        </div>
                                                    </div>
                                                ),
                                            )}
                                        </div>
                                    </div>
                                ))}
                            </div>
                        </div>
                    </div>

                    {/* Where we're heading */}
                    <div className={clsx("w-full")}>
                        <div className={clsx("px-4 landing-md:px-10")}>
                            <h2 className={sectionHeading}>
                                Where we&apos;re{" "}
                                <span className={sectionAccent}>heading</span>.
                            </h2>
                            <p
                                className={clsx(
                                    "mt-4 landing-sm:mt-6",
                                    "max-w-md",
                                    "text-base",
                                    "dark:text-gray-400 text-gray-600",
                                )}
                            >
                                What we&apos;re building right now, and what&apos;s up
                                next.
                            </p>
                        </div>

                        <div
                            className={clsx(
                                "mt-8 landing-sm:mt-12 landing-lg:mt-20",
                                "px-4 landing-md:px-10",
                                "grid",
                                "grid-cols-1 landing-sm:grid-cols-2",
                                "gap-4 landing-sm:gap-6",
                            )}
                        >
                            {inProgress.map((item, index) => (
                                <div key={index} className={cardShell}>
                                    <div className={clsx("flex", "flex-col", "gap-4")}>
                                        <div
                                            className={clsx(
                                                "flex items-center gap-3",
                                            )}
                                        >
                                            <div
                                                className={clsx(
                                                    "text-xl",
                                                    "font-semibold",
                                                    "text-gray-900 dark:text-gray-0",
                                                )}
                                            >
                                                {item.title}
                                            </div>
                                            <div
                                                className={clsx(
                                                    "text-xs",
                                                    "font-semibold",
                                                    "uppercase",
                                                    "tracking-wide",
                                                    "px-2.5 py-1",
                                                    "rounded-full",
                                                    "dark:bg-refine-cyan-alt/10 bg-refine-blue/10",
                                                    "dark:text-refine-cyan-alt text-refine-blue",
                                                )}
                                            >
                                                In progress
                                            </div>
                                        </div>
                                        <div
                                            className={clsx(
                                                "text-base",
                                                "dark:text-gray-400 text-gray-600",
                                            )}
                                        >
                                            {item.description}
                                        </div>
                                    </div>
                                </div>
                            ))}
                        </div>

                        <div
                            className={clsx(
                                "mt-4 landing-sm:mt-6",
                                "px-4 landing-md:px-10",
                                "grid",
                                "grid-cols-1 landing-sm:grid-cols-2",
                                "gap-4",
                            )}
                        >
                            {plannedFeatures.map((item, index) => (
                                <div
                                    key={index}
                                    className={clsx(
                                        "not-prose",
                                        "p-4 landing-sm:p-6",
                                        "flex",
                                        "items-center",
                                        "dark:bg-landing-noise",
                                        "dark:bg-gray-800 bg-gray-50",
                                        "rounded-2xl",
                                    )}
                                >
                                    <div
                                        className={clsx(
                                            "text-base",
                                            "font-semibold",
                                            "text-gray-900 dark:text-gray-0",
                                        )}
                                    >
                                        {item.title}
                                    </div>
                                </div>
                            ))}
                        </div>

                        <div className={clsx("px-4 landing-md:px-10", "mt-6")}>
                            <Link
                                to="/roadmap/"
                                className={clsx(
                                    "text-sm font-semibold",
                                    "dark:text-refine-cyan-alt text-refine-blue",
                                )}
                            >
                                See the full roadmap →
                            </Link>
                        </div>
                    </div>

                    {/* Our story */}
                    <div className={clsx("w-full")}>
                        <div className={clsx("px-4 landing-md:px-10")}>
                            <h2 className={sectionHeading}>
                                Our <span className={sectionAccent}>story</span>.
                            </h2>
                            <p
                                className={clsx(
                                    "mt-4 landing-sm:mt-6",
                                    "max-w-md",
                                    "text-base",
                                    "dark:text-gray-400 text-gray-600",
                                )}
                            >
                                Containers had changed everything about infrastructure -
                                but hosting control panels were still stuck in the past.
                            </p>
                        </div>

                        <div
                            className={clsx(
                                "mt-8 landing-sm:mt-12 landing-lg:mt-20",
                                "px-4 landing-md:px-10",
                                "grid",
                                "grid-cols-1",
                                "gap-4 landing-sm:gap-6",
                            )}
                        >
                            {story.map((item, index) => (
                                <div key={index} className={cardShell}>
                                    <div>{item.icon}</div>
                                    <div className={clsx("flex", "flex-col", "gap-4")}>
                                        <div
                                            className={clsx(
                                                "text-xl",
                                                "font-semibold",
                                                "text-gray-900 dark:text-gray-0",
                                            )}
                                        >
                                            {item.title}
                                        </div>
                                        <div
                                            className={clsx(
                                                "text-base",
                                                "dark:text-gray-400 text-gray-600",
                                            )}
                                        >
                                            {item.description}
                                        </div>
                                    </div>
                                </div>
                            ))}
                        </div>
                    </div>

                    {/* What we believe */}
                    <div className={clsx("w-full")}>
                        <div className={clsx("px-4 landing-md:px-10")}>
                            <h2 className={sectionHeading}>
                                What we <span className={sectionAccent}>believe</span>.
                            </h2>
                            <p
                                className={clsx(
                                    "mt-4 landing-sm:mt-6",
                                    "max-w-md",
                                    "text-base",
                                    "dark:text-gray-400 text-gray-600",
                                )}
                            >
                                The principles that shape every decision we make about
                                OpenPanel.
                            </p>
                        </div>

                        <div
                            className={clsx(
                                "mt-8 landing-sm:mt-12 landing-lg:mt-20",
                                "px-4 landing-md:px-10",
                                "grid",
                                "grid-cols-1 landing-sm:grid-cols-2",
                                "gap-4 landing-sm:gap-6",
                            )}
                        >
                            {beliefs.map((item, index) => (
                                <div key={index} className={cardShell}>
                                    <div>{item.icon}</div>
                                    <div className={clsx("flex", "flex-col", "gap-4")}>
                                        <div
                                            className={clsx(
                                                "text-xl",
                                                "font-semibold",
                                                "text-gray-900 dark:text-gray-0",
                                            )}
                                        >
                                            {item.title}
                                        </div>
                                        <div
                                            className={clsx(
                                                "text-base",
                                                "dark:text-gray-400 text-gray-600",
                                            )}
                                        >
                                            {item.description}
                                        </div>
                                    </div>
                                </div>
                            ))}
                        </div>
                    </div>

                    {/* Independent & fairly priced */}
                    <div className={clsx("px-4 landing-md:px-10")}>
                        <div
                            className={clsx(
                                "not-prose",
                                "flex flex-col items-center text-center",
                                "gap-4",
                                "px-6 py-12 landing-sm:px-20 landing-sm:py-16",
                                "dark:bg-landing-noise",
                                "dark:bg-gray-800 bg-gray-50",
                                "rounded-2xl landing-sm:rounded-3xl",
                            )}
                        >
                            <h2
                                className={clsx(
                                    "text-2xl landing-sm:text-[32px] landing-sm:leading-[40px]",
                                    "tracking-tight",
                                    "font-semibold",
                                    "p-0",
                                    "dark:text-gray-0 text-gray-900",
                                )}
                            >
                                Independent by choice.
                            </h2>
                            <p
                                className={clsx(
                                    "max-w-xl",
                                    "text-base",
                                    "dark:text-gray-400 text-gray-600",
                                )}
                            >
                                Over the years we turned down offers from competitors and
                                venture capitalists. Every one came with conditions that
                                would&apos;ve compromised the product, so we stayed
                                self-funded instead. One promise we stand by:{" "}
                                <span
                                    className={clsx(
                                        "font-semibold",
                                        "dark:text-gray-0 text-gray-900",
                                    )}
                                >
                                    we will never charge per user or per website
                                </span>
                                .
                            </p>
                            <div
                                className={clsx(
                                    "mt-4",
                                    "flex flex-col landing-xs:flex-row",
                                    "items-center justify-center",
                                    "gap-4",
                                    "w-full landing-xs:w-auto",
                                )}
                            >
                                <Link
                                    to="/community/"
                                    className={clsx(
                                        "w-full landing-xs:w-auto",
                                        "!text-gray-0 dark:!text-gray-900",
                                        "bg-refine-blue dark:bg-refine-cyan-alt",
                                        "transition-[filter]",
                                        "duration-150",
                                        "ease-in-out",
                                        "hover:brightness-110",
                                        "hover:!no-underline",
                                        "rounded-3xl",
                                        "py-3 px-8",
                                        "flex items-center justify-center gap-2",
                                        "text-base font-semibold",
                                    )}
                                >
                                    Community Edition
                                </Link>
                                <Link
                                    to="/enterprise"
                                    className={clsx(
                                        "w-full landing-xs:w-auto",
                                        "dark:text-refine-cyan-alt text-refine-blue",
                                        "dark:bg-refine-cyan-alt/10 bg-refine-blue/10",
                                        "transition-[filter]",
                                        "duration-150",
                                        "ease-in-out",
                                        "hover:brightness-110",
                                        "hover:!no-underline",
                                        "rounded-3xl",
                                        "py-3 px-8",
                                        "flex items-center justify-center gap-2",
                                        "text-base font-semibold",
                                    )}
                                >
                                    Enterprise Edition
                                </Link>
                            </div>
                        </div>
                    </div>

                    {/* Team */}
                    <div className={clsx("w-full")}>
                        <div className={clsx("px-4 landing-md:px-10")}>
                            <h2 className={sectionHeading}>
                                Who&apos;s <span className={sectionAccent}>behind it</span>.
                            </h2>
                            <p
                                className={clsx(
                                    "mt-4 landing-sm:mt-6",
                                    "max-w-md",
                                    "text-base",
                                    "dark:text-gray-400 text-gray-600",
                                )}
                            >
                                A small, independent team building OpenPanel full-time.
                            </p>
                        </div>

                        <div
                            className={clsx(
                                "mt-8 landing-sm:mt-12",
                                "px-4 landing-md:px-10",
                                "grid",
                                "grid-cols-2 landing-sm:grid-cols-4",
                                "gap-4",
                            )}
                        >
                            {team.map(({ name, avatar, role1, role2, linkedin }) => {
                                const isLocalImage = avatar.startsWith("/");
                                return (
                                    <Link
                                        key={name}
                                        to={linkedin}
                                        target="_blank"
                                        rel="noopener noreferrer"
                                        className={clsx(
                                            "not-prose",
                                            "flex flex-col items-center text-center",
                                            "gap-3",
                                            "p-4 landing-sm:p-6",
                                            "dark:bg-landing-noise",
                                            "dark:bg-gray-800 bg-gray-50",
                                            "rounded-2xl",
                                            "transition-[filter]",
                                            "duration-150",
                                            "ease-in-out",
                                            "hover:brightness-110",
                                            "hover:!no-underline",
                                        )}
                                    >
                                        {isLocalImage ? (
                                            <img
                                                src={avatar}
                                                alt={name}
                                                className={clsx(
                                                    "w-16 h-16 landing-sm:w-20 landing-sm:h-20",
                                                    "rounded-full",
                                                    "object-cover",
                                                    "m-0",
                                                )}
                                            />
                                        ) : (
                                            <div
                                                className={clsx(
                                                    "w-16 h-16 landing-sm:w-20 landing-sm:h-20",
                                                    "rounded-full",
                                                    "flex items-center justify-center",
                                                    "bg-gradient-to-br from-refine-cyan-alt to-refine-blue",
                                                    "text-gray-0",
                                                    "text-lg landing-sm:text-xl",
                                                    "font-semibold",
                                                )}
                                            >
                                                {getInitials(name)}
                                            </div>
                                        )}
                                        <div>
                                            <div
                                                className={clsx(
                                                    "text-sm landing-sm:text-base",
                                                    "font-semibold",
                                                    "text-gray-900 dark:text-gray-0",
                                                )}
                                            >
                                                {name}
                                            </div>
                                            <div
                                                className={clsx(
                                                    "text-xs landing-sm:text-sm",
                                                    "text-gray-500 dark:text-gray-400",
                                                )}
                                            >
                                                {role1}
                                            </div>
                                            {role2 && (
                                                <div
                                                    className={clsx(
                                                        "text-xs landing-sm:text-sm",
                                                        "text-gray-500 dark:text-gray-400",
                                                    )}
                                                >
                                                    {role2}
                                                </div>
                                            )}
                                        </div>
                                    </Link>
                                );
                            })}
                        </div>
                    </div>

                    {/* Bottom CTA */}
                    <div className={clsx("px-4 landing-md:px-10")}>
                        <div
                            className={clsx(
                                "not-prose",
                                "flex flex-col items-center text-center",
                                "gap-4",
                                "px-6 py-12 landing-sm:px-20 landing-sm:py-16",
                                "dark:bg-landing-noise",
                                "dark:bg-gray-800 bg-gray-50",
                                "rounded-2xl landing-sm:rounded-3xl",
                            )}
                        >
                            <h2
                                className={clsx(
                                    "text-2xl landing-sm:text-[32px] landing-sm:leading-[40px]",
                                    "tracking-tight",
                                    "font-semibold",
                                    "p-0",
                                    "dark:text-gray-0 text-gray-900",
                                )}
                            >
                                Come build the future of hosting with us.
                            </h2>
                            <p
                                className={clsx(
                                    "max-w-lg",
                                    "text-base",
                                    "dark:text-gray-400 text-gray-600",
                                )}
                            >
                                Join our community, contribute on GitHub, or just say
                                hello - we&apos;re glad you&apos;re here.
                            </p>
                            <div
                                className={clsx(
                                    "mt-4",
                                    "flex flex-col landing-xs:flex-row",
                                    "items-center justify-center",
                                    "gap-4",
                                    "w-full landing-xs:w-auto",
                                )}
                            >
                                <Link
                                    to="https://discord.com/invite/7bNY8fANqF"
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    className={clsx(
                                        "w-full landing-xs:w-auto",
                                        "!text-gray-0 dark:!text-gray-900",
                                        "bg-refine-blue dark:bg-refine-cyan-alt",
                                        "transition-[filter]",
                                        "duration-150",
                                        "ease-in-out",
                                        "hover:brightness-110",
                                        "hover:!no-underline",
                                        "rounded-3xl",
                                        "py-3 px-8",
                                        "flex items-center justify-center gap-2",
                                        "text-base font-semibold",
                                    )}
                                >
                                    <HeaderDiscordIcon width={20} height={20} />
                                    Join Discord
                                </Link>
                                <Link
                                    to="https://github.com/stefanpejcic/openpanel"
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    className={clsx(
                                        "w-full landing-xs:w-auto",
                                        "dark:text-refine-cyan-alt text-refine-blue",
                                        "dark:bg-refine-cyan-alt/10 bg-refine-blue/10",
                                        "transition-[filter]",
                                        "duration-150",
                                        "ease-in-out",
                                        "hover:brightness-110",
                                        "hover:!no-underline",
                                        "rounded-3xl",
                                        "py-3 px-8",
                                        "flex items-center justify-center gap-2",
                                        "text-base font-semibold",
                                    )}
                                >
                                    <FooterGithubIcon width={20} height={20} />
                                    GitHub
                                </Link>
                            </div>
                            <Link
                                to="mailto:info@openpanel.com"
                                className={clsx(
                                    "mt-2",
                                    "flex items-center gap-2",
                                    "text-sm",
                                    "dark:text-gray-400 text-gray-600",
                                    "hover:!no-underline",
                                )}
                            >
                                <MailIcon width={16} height={14} />
                                info@openpanel.com
                            </Link>
                        </div>
                    </div>
                </div>

                <BlogFooter />
            </div>
        </CommonLayout>
    );
};

export default About;
