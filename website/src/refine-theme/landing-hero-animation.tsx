import clsx from "clsx";
import { useInView } from "framer-motion";
import React from "react";
import {
    LandingHeroBeamGlowSvg,
    LandingHeroBeamSvg,
} from "./icons/landing-hero-beam";
import { LandingHeroCenterSvg } from "./icons/landing-hero-center";
import { LandingHeroGridSvg } from "./icons/landing-hero-grid";
import { LandingHeroRestApiIcon } from "./icons/landing-hero/rest-api";
import { LandingHeroMcpIcon } from "./icons/landing-hero/mcp";
import { LandingHeroWordPressIcon } from "./icons/landing-hero/wordpress";
import { LandingHeroPythonIcon } from "./icons/landing-hero/python";
import { LandingHeroNodeJSIcon } from "./icons/landing-hero/nodejs";
import { LandingHeroWebsiteBuilderIcon } from "./icons/landing-hero/website-builder";
import { LandingHeroContainersIcon } from "./icons/landing-hero/containers";
import { LandingHeroCaddyIcon } from "./icons/landing-hero/caddy";
import { LandingHeroOpenLiteSpeedIcon } from "./icons/landing-hero/openlitespeed";
import { LandingHeroApacheIcon } from "./icons/landing-hero/apache";
import { LandingHeroNginxIcon } from "./icons/landing-hero/nginx";
import { LandingHeroMariaDBIcon } from "./icons/landing-hero/mariadb";
import { LandingHeroPostgreSQLIcon } from "./icons/landing-hero/postgresql";
import { LandingHeroMySQLIcon } from "./icons/landing-hero/mysql";
import { LandingHeroPerconaIcon } from "./icons/landing-hero/percona";
import { LandingHeroLlmsTxtIcon } from "./icons/landing-hero/llms-txt";
import { LandingHeroTailwindCSSIcon } from "./icons/landing-hero/tailwindcss";
import { LandingHeroAlpineJSIcon } from "./icons/landing-hero/alpinejs";
import { LandingHeroTremorIcon } from "./icons/landing-hero/tremor";
import { LandingHeroGoIcon } from "./icons/landing-hero/go";
import { LandingHeroBashIcon } from "./icons/landing-hero/bash";
import { LandingHeroVarnishIcon } from "./icons/landing-hero/varnish";
import { LandingHeroMemcachedIcon } from "./icons/landing-hero/memcached";
import { LandingHeroRedisIcon } from "./icons/landing-hero/redis";
import { LandingHeroValkeyIcon } from "./icons/landing-hero/valkey";
import { LandingHeroWHMCSIcon } from "./icons/landing-hero/whmcs";
import { LandingHeroBlestaIcon } from "./icons/landing-hero/blesta";
import { LandingHeroFossBillingIcon } from "./icons/landing-hero/fossbilling";
import { LandingHeroPaymenterIcon } from "./icons/landing-hero/paymenter";
import { LandingHeroPasskeysIcon } from "./icons/landing-hero/passkeys";
import { LandingHeroTwoFaIcon } from "./icons/landing-hero/two-fa";
import { LandingHeroPasswordsIcon } from "./icons/landing-hero/passwords";
import { LandingHeroAnimationItem } from "./landing-hero-animation-item";

type ItemType = {
    icon: React.ComponentType<React.SVGProps<SVGSVGElement>>;
    name: string;
    color: string;
    rayClassName?: string;
};

type GroupOption = {
    section: string;
    items: ItemType[];
};

const mono = (props: React.SVGProps<SVGSVGElement>) =>
    clsx("text-gray-1000 dark:text-gray-0", props.className);

// ---- combo A groups ----

const aiFeatures: GroupOption = {
    section: "AI features",
    items: [
        {
            name: "MCP",
            icon: (props) => (
                <LandingHeroMcpIcon
                    {...props}
                    className={mono(props)}
                    style={{ marginLeft: "1.5em" }}
                />
            ),
            color: "#8B5CF6",
        },
        {
            name: "llms.txt",
            icon: (props) => (
                <LandingHeroLlmsTxtIcon
                    {...props}
                    className={mono(props)}
                    style={{ marginLeft: "1.5em" }}
                />
            ),
            color: "#64748B",
        },
        {
            name: "REST API",
            icon: (props) => (
                <LandingHeroRestApiIcon
                    {...props}
                    className={mono(props)}
                    style={{ marginLeft: "1.5em" }}
                />
            ),
            color: "#3B82F6",
        },
    ],
};

const frontend: GroupOption = {
    section: "Frontend",
    items: [
        {
            name: "TailwindCSS",
            icon: (props) => (
                <LandingHeroTailwindCSSIcon
                    {...props}
                    style={{ marginLeft: "1.5em" }}
                />
            ),
            color: "#38BDF8",
        },
        {
            name: "AlpineJS",
            icon: (props) => (
                <LandingHeroAlpineJSIcon
                    {...props}
                    style={{ marginLeft: "1.5em" }}
                />
            ),
            color: "#77C1D2",
        },
        {
            name: "TremorUI",
            icon: (props) => (
                <LandingHeroTremorIcon {...props} style={{ marginLeft: "1.5em" }} />
            ),
            color: "#6366F1",
        },
    ],
};

const backend: GroupOption = {
    section: "Backend",
    items: [
        {
            name: "Go",
            icon: (props) => (
                <LandingHeroGoIcon {...props} style={{ marginLeft: "1.5em" }} />
            ),
            color: "#00ACD7",
        },
        {
            name: "Bash",
            icon: (props) => (
                <LandingHeroBashIcon {...props} style={{ marginLeft: "1.5em" }} />
            ),
            color: "#4EAA25",
        },
    ],
};

const billingIntegrations: GroupOption = {
    section: "Billing",
    items: [
        {
            name: "WHMCS",
            icon: (props) => (
                <LandingHeroWHMCSIcon {...props} style={{ marginLeft: "1.5em" }} />
            ),
            color: "#3CB371",
        },
        {
            name: "Blesta",
            icon: (props) => (
                <LandingHeroBlestaIcon {...props} style={{ marginLeft: "1.5em" }} />
            ),
            color: "#6AB31D",
        },
        {
            name: "FOSSBilling",
            icon: (props) => (
                <LandingHeroFossBillingIcon
                    {...props}
                    style={{ marginLeft: "1.5em" }}
                />
            ),
            color: "#22C55E",
        },
        {
            name: "paymenter.org",
            icon: (props) => (
                <LandingHeroPaymenterIcon
                    {...props}
                    style={{ marginLeft: "1.5em" }}
                />
            ),
            color: "#4060FF",
        },
    ],
};

const authentication: GroupOption = {
    section: "authentication",
    items: [
        {
            name: "Passkeys",
            icon: (props) => (
                <LandingHeroPasskeysIcon
                    {...props}
                    className={mono(props)}
                    style={{ marginLeft: "1.5em" }}
                />
            ),
            color: "#8B5CF6",
        },
        {
            name: "2FA",
            icon: (props) => (
                <LandingHeroTwoFaIcon
                    {...props}
                    className={mono(props)}
                    style={{ marginLeft: "1.5em" }}
                />
            ),
            color: "#EC4899",
        },
        {
            name: "REST API",
            icon: (props) => (
                <LandingHeroRestApiIcon
                    {...props}
                    className={mono(props)}
                    style={{ marginLeft: "1.5em" }}
                />
            ),
            color: "#3B82F6",
        },
        {
            name: "Passwords",
            icon: (props) => (
                <LandingHeroPasswordsIcon
                    {...props}
                    className={mono(props)}
                    style={{ marginLeft: "1.5em" }}
                />
            ),
            color: "#F59E0B",
        },
    ],
};

// ---- combo B groups ----

const applications: GroupOption = {
    section: "applications",
    items: [
        {
            name: "WordPress",
            icon: (props) => (
                <LandingHeroWordPressIcon
                    {...props}
                    style={{ marginLeft: "1.5em" }}
                />
            ),
            color: "#21759B",
        },
        {
            name: "Python",
            icon: (props) => (
                <LandingHeroPythonIcon {...props} style={{ marginLeft: "1.5em" }} />
            ),
            color: "#3776AB",
        },
        {
            name: "NodeJS",
            icon: (props) => (
                <LandingHeroNodeJSIcon {...props} style={{ marginLeft: "1.5em" }} />
            ),
            color: "#339933",
        },
        {
            name: "Website Builder",
            icon: (props) => (
                <LandingHeroWebsiteBuilderIcon
                    {...props}
                    className={mono(props)}
                    style={{ marginLeft: "1.5em" }}
                />
            ),
            color: "#F59E0B",
        },
        {
            name: "Containers",
            icon: (props) => (
                <LandingHeroContainersIcon
                    {...props}
                    className={mono(props)}
                    style={{ marginLeft: "1.5em" }}
                />
            ),
            color: "#2496ED",
        },
    ],
};

const caching: GroupOption = {
    section: "Caching",
    items: [
        {
            name: "Varnish",
            icon: (props) => (
                <LandingHeroVarnishIcon {...props} style={{ marginLeft: "1.5em" }} />
            ),
            color: "#0072BC",
        },
        {
            name: "Memcached",
            icon: (props) => (
                <LandingHeroMemcachedIcon
                    {...props}
                    style={{ marginLeft: "1.5em" }}
                />
            ),
            color: "#C83737",
        },
        {
            name: "Redis",
            icon: (props) => (
                <LandingHeroRedisIcon {...props} style={{ marginLeft: "1.5em" }} />
            ),
            color: "#DC382D",
        },
        {
            name: "Valkey",
            icon: (props) => (
                <LandingHeroValkeyIcon {...props} style={{ marginLeft: "1.5em" }} />
            ),
            color: "#FFD43B",
        },
    ],
};

const webservers: GroupOption = {
    section: "webservers",
    items: [
        {
            name: "Caddy",
            icon: (props) => (
                <LandingHeroCaddyIcon {...props} style={{ marginLeft: "1.5em" }} />
            ),
            color: "#1F88C0",
        },
        {
            name: "OpenLiteSpeed",
            icon: (props) => (
                <LandingHeroOpenLiteSpeedIcon
                    {...props}
                    style={{ marginLeft: "1.5em" }}
                />
            ),
            color: "#75D1DF",
        },
        {
            name: "Apache",
            icon: (props) => (
                <LandingHeroApacheIcon {...props} style={{ marginLeft: "1.5em" }} />
            ),
            color: "#D22128",
        },
        {
            name: "Nginx",
            icon: (props) => (
                <LandingHeroNginxIcon {...props} style={{ marginLeft: "1.5em" }} />
            ),
            color: "#009639",
        },
    ],
};

const databases: GroupOption = {
    section: "databases",
    items: [
        {
            name: "MariaDB",
            icon: (props) => (
                <LandingHeroMariaDBIcon {...props} style={{ marginLeft: "1.5em" }} />
            ),
            color: "#1F7A8C",
        },
        {
            name: "PostgreSQL",
            icon: (props) => (
                <LandingHeroPostgreSQLIcon
                    {...props}
                    style={{ marginLeft: "1.5em" }}
                />
            ),
            color: "#336791",
        },
        {
            name: "MySQL",
            icon: (props) => (
                <LandingHeroMySQLIcon {...props} style={{ marginLeft: "1.5em" }} />
            ),
            color: "#00758F",
        },
        {
            name: "Percona",
            icon: (props) => (
                <LandingHeroPerconaIcon {...props} style={{ marginLeft: "1.5em" }} />
            ),
            color: "#FF6D01",
        },
    ],
};

// The whole animation shows one of exactly two combos, chosen once per page
// load — not 4 independently-randomized corners. Combo A's position 2 has
// its own further random choice between AI features and Billing
// Integrations. Starts on combo B (the original/default layout) so server
// and client render the same markup on hydration, then randomizes
// client-side only — same pattern as HERO_VARIANTS in
// landing-hero-section.tsx.
function usePositions() {
    const [combo, setCombo] = React.useState<"A" | "B">("B");
    const [comboAPosition2, setComboAPosition2] = React.useState<
        "ai" | "billing"
    >("ai");

    React.useEffect(() => {
        const chosenCombo = Math.random() < 0.5 ? "A" : "B";
        setCombo(chosenCombo);
        if (chosenCombo === "A") {
            setComboAPosition2(Math.random() < 0.5 ? "ai" : "billing");
        }
    }, []);

    if (combo === "A") {
        return {
            position1: frontend,
            position2: backend,
            position3: comboAPosition2 === "ai" ? aiFeatures : billingIntegrations,
            position4: authentication,
        };
    }

    return {
        position1: applications,
        position2: caching,
        position3: webservers,
        position4: databases,
    };
}

export const LandingHeroAnimation = React.memo(function HeroAnimation() {
    const ref = React.useRef<HTMLDivElement>(null);
    const inView = useInView(ref);

    const { position1, position2, position3, position4 } = usePositions();

    const [active1, setActive1] = React.useState(0);
    const [active2, setActive2] = React.useState(0);
    const [active3, setActive3] = React.useState(0);
    const [active4, setActive4] = React.useState(0);

    React.useEffect(() => {
        if (inView) {
            let t1: NodeJS.Timeout | null = null;
            let t2: NodeJS.Timeout | null = null;
            let t3: NodeJS.Timeout | null = null;

            const interval = setInterval(() => {
                if (t1) clearTimeout(t1);
                if (t2) clearTimeout(t2);
                if (t3) clearTimeout(t3);

                setActive1((prev) => (prev + 1) % position1.items.length);
                t1 = setTimeout(() => {
                    setActive2((prev) => (prev + 1) % position2.items.length);
                }, 2000);
                t2 = setTimeout(() => {
                    setActive3((prev) => (prev + 1) % position3.items.length);
                }, 4000);
                t3 = setTimeout(() => {
                    setActive4((prev) => (prev + 1) % position4.items.length);
                }, 6000);
            }, 8000);

            return () => {
                clearInterval(interval);
                if (t1) clearTimeout(t1);
                if (t2) clearTimeout(t2);
                if (t3) clearTimeout(t3);
            };
        }
    }, [inView, position1, position2, position3, position4]);

    return (
        <div ref={ref} className={clsx()}>
            <div className={clsx("relative", "w-min")}>
                <LandingHeroGridSvg
                    className={clsx(
                        "w-[404px]",
                        "h-[360px]",
                        "landing-lg:w-[690px]",
                        "landing-lg:h-[480px]",
                        "left-0",
                        "top-0",
                        "bottom-0",
                        "right-0",
                    )}
                />
                <LandingHeroCenterSvg
                    className={clsx(
                        "absolute",
                        "left-1/2 top-1/2",
                        "-translate-x-1/2 -translate-y-1/2",
                        "z-[1]",
                    )}
                />
                <div
                    className={clsx(
                        "hidden",
                        "landing-lg:flex",
                        "absolute",
                        "left-0",
                        "top-0",
                        "bottom-0",
                        "right-0",
                        "w-full",
                        "h-full",
                        "py-12",
                        "px-[89px]",
                        "flex-col",
                        "items-start",
                        "justify-between",
                    )}
                >
                    <div
                        className={clsx(
                            "w-full",
                            "flex",
                            "items-start",
                            "justify-between",
                        )}
                    >
                        <LandingHeroAnimationItem
                            vertical="top"
                            horizontal="left"
                            section={position1.section}
                            {...position1.items[active1]}
                            previousName={
                                position1.items[
                                    (active1 - 1 + position1.items.length) %
                                        position1.items.length
                                ].name ?? position1.items[active1].name
                            }
                        />
                        <LandingHeroAnimationItem
                            vertical="top"
                            horizontal="right"
                            section={position2.section}
                            {...position2.items[active2]}
                            previousName={
                                position2.items[
                                    (active2 - 1 + position2.items.length) %
                                        position2.items.length
                                ].name ?? position2.items[active2].name
                            }
                        />
                    </div>
                    <div
                        className={clsx(
                            "mt-auto",
                            "w-full",
                            "flex",
                            "items-end",
                            "justify-between",
                        )}
                    >
                        <LandingHeroAnimationItem
                            vertical="bottom"
                            horizontal="left"
                            section={position3.section}
                            {...position3.items[active3]}
                            previousName={
                                position3.items[
                                    (active3 - 1 + position3.items.length) %
                                        position3.items.length
                                ].name ?? position3.items[active3].name
                            }
                        />
                        <LandingHeroAnimationItem
                            vertical="bottom"
                            horizontal="right"
                            section={position4.section}
                            {...position4.items[active4]}
                            previousName={
                                position4.items[
                                    (active4 - 1 + position4.items.length) %
                                        position4.items.length
                                ].name ?? position4.items[active4].name
                            }
                        />
                    </div>
                </div>
            </div>
        </div>
    );
});

// Rendered as a separate, independently bottom-anchored sibling of
// LandingHeroAnimation (see landing-hero-section.tsx) so the connecting
// ray/glow tracks down to the row's actual bottom edge — and stays flush
// against the section below — regardless of how tall the row grows from a
// longer tagline. The icon grid above it stays fixed/top-anchored and never
// moves; only this beam needs to reach a variable-length gap.
export const LandingHeroAnimationBeam = () => (
    <div className={clsx("relative", "w-full", "h-64")}>
        <LandingHeroBeamSvg
            className={clsx(
                "z-[0]",
                "absolute",
                "left-1/2",
                "bottom-0",
                "-translate-x-1/2",
                "dark:animate-landing-hero-beam-line",
            )}
        />
        <LandingHeroBeamGlowSvg
            className={clsx(
                "z-[0]",
                "absolute",
                "left-1/2",
                "bottom-0",
                "-translate-x-1/2",
                "blur-sm",
                "dark:animate-landing-hero-beam-glow",
            )}
            style={{
                fillOpacity: 0,
                filter: "drop-shadow(rgba(71, 235, 235,0.1) 0px 0px 0px) drop-shadow(rgba(71, 235, 235,0.15) 0px 0px 10px)",
            }}
        />
        <div
            className={clsx(
                "-mb-px",
                "overflow-hidden",
                "absolute",
                "left-1/2",
                "-translate-x-1/2",
                "bottom-0",
                "z-[1]",
            )}
        >
            <div
                className={clsx(
                    "relative",
                    "w-40",
                    "h-px",
                    "bg-landing-hero-beam-bottom-light dark:bg-landing-hero-beam-bottom",
                    "animate-landing-hero-beam-bottom",
                )}
            ></div>
        </div>
    </div>
);
