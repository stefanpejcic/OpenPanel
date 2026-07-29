import React from "react";
import clsx from "clsx";
import { LandingHeroGithubStars } from "./landing-hero-github-stars";
import { LandingStartActionIcon } from "./icons/landing-start-action";

import { LandingHeroAnimation } from "./landing-hero-animation";
import { LandingCopyCommandButton } from "./landing-copy-command-button";
import Link from "@docusaurus/Link";
import { LandingHeroShowcaseSection } from "./landing-hero-showcase-section";

// Heading and tagline rotate together as pairs so the copy always
// reads as one coherent message. Taglines are kept at matching
// length/word-count on purpose — the hero glow image is positioned
// against this block's height, so a longer/shorter variant shifts the
// layout and misaligns it.
const HERO_VARIANTS = [
    // technical / containers — default, rendered on the server
    {
        heading: "Next Generation",
        tagline:
            "OpenPanel is a multi-user web hosting panel designed around lightweight containers. Each user gets their own fully isolated environment, complete with a separate MySQL server, PHP version, Redis instance, and full root access.",
    },
    // open-source
    {
        heading: "Open Source",
        tagline:
            "OpenPanel is open source and community-driven, and you can self-host it anywhere you like, on your own hardware or any cloud provider. No per-account fees, no vendor lock-in — containers do the isolation work for you, at no cost at all, ever.",
    },
    // competitive vs cPanel/Plesk
    {
        heading: "No Lock-In",
        tagline:
            "Tired of cPanel or Plesk license fees eating into your margins? OpenPanel gives you the same multi-user hosting workflow you already know, built on containers instead of legacy shared Linux accounts, and it costs you nothing, ever.",
    },
];

export const LandingHeroSection = ({ className }: { className?: string }) => {
    const [variant, setVariant] = React.useState(HERO_VARIANTS[0]);

    React.useEffect(() => {
        setVariant(
            HERO_VARIANTS[Math.floor(Math.random() * HERO_VARIANTS.length)],
        );
    }, []);

    return (
        <div
            className={clsx(
                "flex",
                "flex-col",
                "w-full",
                "gap-2",
                className,
            )}
        >
            <div
                className={clsx(
                    "px-2 landing-sm:px-0",
                    "flex",
                    "flex",
                    "w-full",
                    "relative",
                    "min-h-[360px]",
                    "landing-lg:min-h-[480px]",
                    "py-4",
                )}
            >
                <div
                    className={clsx(
                        "landing-sm:pl-10",
                        "flex",
                        "flex-col",
                        "justify-center",
                        "gap-6",
                        "z-[1]",
                        "landing-lg:justify-between",
                        "landing-lg:py-8",
                    )}
                >
                    <LandingHeroGithubStars />
                    <div className={clsx("flex", "flex-col", "gap-6")}>
                        <h1
                            className={clsx(
                                "text-[32px] leading-[40px]",
                                "tracking-[-0.5%]",
                                "landing-sm:text-[56px] landing-sm:leading-[72px]",
                                "landing-sm:max-w-[588px]",
                                "landing-sm:tracking-[-2%]",
                                "font-extrabold",
                                "text-gray-900 dark:text-gray-0",
                            )}
                        >
			<span class="text-transparent bg-clip-text bg-gradient-to-r text-gradient-to-r from-[#0FBDBD] to-[#26D97F]">{variant.heading}</span><br />Hosting Panel
                        </h1>
                        <p
                            className={clsx(
                                "font-normal",
                                "text-base",
                                "text-gray-600 dark:text-gray-300",
                                "landing-xs:max-w-[384px]",
                            )}
                        >
							{variant.tagline}
						</p>
                        <p
                            className={clsx(
                                "font-normal",
                                "text-base",
                                "text-gray-600 dark:text-gray-300",
                                "landing-xs:max-w-[384px]",
								"hidden",
								"landing-sm:block",
                            )}
                        >
							Try the free Community Edition with command:
						</p>
                    </div>
                    <div
                        className={clsx(
                            "flex",
                            "items-center",
                            "justify-start",
                            "gap-4",
                            "landing-lg:gap-6",
                        )}
                    >
			<LandingCopyCommandButton
			className={clsx("hidden", "landing-sm:block")}
			/>
                    </div>
                </div>
                <div
                    className={clsx(
                        "hidden landing-md:block",
                        "absolute",
                        "top-0",
                        "right-0",
                    )}
                >
                    <LandingHeroAnimation />
                </div>
            </div>
            <LandingHeroShowcaseSection />
        </div>
    );
};
