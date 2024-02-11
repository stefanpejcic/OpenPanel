import React from "react";
import clsx from "clsx";
import { LandingArrowRightIcon } from "./icons/landing-arrow-right";
import { ShowcaseWrapper } from "../components/landing/showcase-wrapper";

const ShowcaseCRM = ({ className }: { className?: string }) => {
    return (
        <ShowcaseWrapper
            className={className}
            render="/img/panel_cropped.png"
            highlights={[]}
        />
    );
};

const ShowcaseHR = ({ className }: { className?: string }) => {
    return (
        <ShowcaseWrapper
            className={className}
            render="/img/openadmin.png"
            highlights={[]}
        />
    );
};

const ShowcaseDevOps = ({ className }: { className?: string }) => {
    return (
        <ShowcaseWrapper
            className={className}
            dark
            render="/img/opencli_preview.png"
            highlights={[]}
        />
    );
};

export const LandingHeroShowcaseSection = ({}) => {
    const [activeApp, setActiveApp] = React.useState(apps[0]);

    const ShowcaseComponent = React.useMemo(() => {
        return activeApp.showcase;
    }, [activeApp.name]);

    return (
        <div
            className={clsx(
                "bg-gray-50 dark:bg-gray-800",
                "flex",
                "flex-col",
                "w-full",
                "rounded-2xl landing-sm:rounded-[32px]",
                "gap-2 landing-sm:gap-4",
                "p-2 landing-sm:p-4",
                "relative",
                "group/showcase",
                "landing-lg:overflow-hidden",
            )}
            style={{ marginTop: "-7px" }}
        >
            <div className={clsx("flex", "w-full", "gap-2")}>
                <div
                    className={clsx(
                        "rounded-3xl",
                        "overflow-y-auto",
                        "flex",
                        "w-full",
                        "gap-2",
                        "scrollbar-hidden",
                        "snap snap-x snap-mandatory",
                        "snap-mandatory",
                    )}
                >
                    <div
                        className={clsx(
                            "rounded-3xl",
                            "flex",
                            "w-auto",
                            "landing-lg:w-full",
                            "items-center",
                            "justify-start",
                            "gap-2",
                            "relative",
                            "bg-gray-0 dark:bg-gray-900",
                        )}
                    >
                        <div
                            className={clsx(
                                "hidden landing-sm:block",
                                "flex-1",
                                "rounded-3xl",
                                "h-full",
                                "bg-gray-200 dark:bg-gray-700",
                                "absolute",
                                "left-0",
                                "top-0",
                                "transition-transform",
                                "duration-150",
                                "ease-out",
                            )}
                            style={{
                                width: `calc(100% / 3)`,
                                minWidth: "244px",
                                transform: `translateX(calc((100% + 8px) * ${apps.findIndex(
                                    (f) => f.name === activeApp.name,
                                )})) translateZ(0px)`,
                            }}
                        />
                        {apps.map((app, index) => (
                            <button
                                key={app.name}
                                type="button"
                                onClick={(event) => {
                                    setActiveApp(app);
                                    // if index i >= 2
                                    // then scroll to the right
                                    event.currentTarget.parentElement?.parentElement?.scrollTo(
                                        {
                                            left:
                                                index >= 2
                                                    ? index * (244 + 8)
                                                    : 0,
                                            behavior: "smooth",
                                        },
                                    );
                                }}
                                className={clsx(
                                    "z-[1]",
                                    "snap-start",
                                    "appearance-none",
                                    "focus:outline-none",
                                    "border-none",
                                    "flex-1",
                                    "break-keep",
                                    "whitespace-nowrap",
                                    "landing-sm:min-w-[244px]",
                                    "py-2",
                                    "landing-sm:py-3.5",
                                    "px-4",
                                    "rounded-3xl",
                                    "transition-colors",
                                    "ease-in-out",
                                    "duration-150",
                                    activeApp.name !== app.name &&
                                        "bg-transparent",
                                    activeApp.name === app.name &&
                                        "bg-gray-200 dark:bg-gray-700",
                                    activeApp.name !== app.name &&
                                        "text-gray-600 dark:text-gray-400",
                                    activeApp.name === app.name &&
                                        "text-gray-900 dark:text-gray-0",
                                    "landing-sm:bg-transparent",
                                    "dark:landing-sm:bg-transparent",
                                    "transition-colors",
                                    "duration-150",
                                    "ease-out",
                                    "text-xs",
                                    "landing-sm:text-sm",
                                )}
                            >
                                {app.name}
                            </button>
                        ))}
                    </div>
                </div>
            </div>
            <div
                className={clsx(
                    "rounded-lg",
                    "landing-md:rounded-xl",
                    "landing-lg:rounded-2xl",
                    "overflow-hidden",
                    "shadow-sm shadow-gray-200 dark:shadow-none",
                    "relative",
                    "group/showcase-inner",
                )}
            >
                <div
                    className={clsx(
                        "w-full",
                        "h-auto",
                        "aspect-[1168/736]",
                        "transition-colors",
                        "duration-150",
                        "ease-in-out",
                        activeApp.dark ? "bg-gray-900" : "bg-gray-0",
                    )}
                />
                <ShowcaseComponent
                    className={clsx(
                        "animate-showcase-reveal",
                        "absolute",
                        "left-0",
                        "top-0",
                        "w-full",
                        "rounded-lg",
                        "landing-md:rounded-xl",
                        "landing-lg:rounded-2xl",
                        "overflow-hidden",
                    )}
                />
                <div
                    key={activeApp.name}
                    className={clsx(
                        "hidden",
                        "landing-lg:block",
                        "landing-lg:opacity-0",
                        "landing-lg:translate-y-24",
                        "landing-lg:group-hover/showcase-inner:opacity-100 landing-lg:group-hover/showcase-inner:translate-y-0",
                        "duration-300",
                        "ease-in-out",
                        "transition-[opacity,transform,background-color,color]",
                        "absolute",
                        "left-0",
                        "bottom-0",
                        "right-0",
                        "w-full",
                        "h-24",
                        "opacity-0",
                        activeApp.dark &&
                            "bg-[linear-gradient(0deg,_#14141F_30%,_transparent_90%,_transparent_100%)]",
                        !activeApp.dark &&
                            "bg-[linear-gradient(0deg,_#FFFFFF_30%,_transparent_90%,_transparent_100%)]",
                        "rounded-bl-lg rounded-br-lg",
                        "landing-md:rounded-bl-xl landing-md:rounded-br-xl",
                        "landing-lg:rounded-bl-2xl landing-lg:rounded-br-2xl",
                    )}
                />
                <div
                    className={clsx(
                        "flex",
                        "items-center",
                        "justify-center",
                        "landing-lg:-mb-4",
                    )}
                >
                    <a
                        href={activeApp.link}
                        target="_blank"
                        rel="noopener noreferrer"
                        className={clsx(
                            "hidden",
                            "landing-lg:flex",
                            "landing-lg:opacity-0",
                            "landing-lg:translate-y-8",
                            "landing-lg:group-hover/showcase-inner:opacity-100 landing-lg:group-hover/showcase-inner:translate-y-0",
                            "duration-150",
                            "delay-75",
                            "ease-in-out",
                            "transition-all",
                            "landing-lg:mt-[-144px]",
                            "hover:no-underline",
                            "z-[3]",
                            "py-2 landing-sm:py-4",
                            "pl-4 pr-4 landing-sm:pl-6 landing-sm:pr-4",
                            "rounded-[32px] landing-sm:rounded-[48px]",
                            "items-center",
                            "justify-center",
                            "gap-2",
                            "bg-refine-blue dark:bg-refine-cyan-alt",
                            "bg-opacity-10 dark:bg-opacity-10",
                            "landing-lg:bg-opacity-100 dark:landing-lg:bg-opacity-100",
                            "text-refine-blue dark:text-refine-cyan-alt",
                            "landing-lg:text-gray-0 dark:landing-lg:text-gray-900",
                            "hover:brightness-125",
                            "landing-lg:hover:scale-105 landing-lg:hover:brightness-100",
                            "hover:text-refine-blue dark:hover:text-refine-cyan-alt",
                            "landing-lg:hover:text-gray-0 dark:landing-lg:hover:text-gray-900",
                            "landing-lg:border-8 landing-lg:border-solid",
                            activeApp.dark
                                ? "landing-lg:border-gray-900"
                                : "landing-lg:border-gray-0",
                        )}
                    >
                        <span
                            className={clsx(
                                "text-xs landing-sm:text-base",
                                "font-semibold",
                            )}
                        >
                            {activeApp.label}
                        </span>
                        <LandingArrowRightIcon />
                    </a>
                </div>
            </div>
            <div
                className={clsx(
                    "flex",
                    "items-center",
                    "justify-center",
                    "landing-lg:-mb-4",
                )}
            >
                <a
                    href={activeApp.link}
                    target="_blank"
                    rel="noopener noreferrer"
                    className={clsx(
                        "landing-lg:opacity-0",
                        "duration-150",
                        "delay-75",
                        "ease-in-out",
                        "transition-all",
                        "hover:no-underline",
                        "z-[3]",
                        "py-2 landing-sm:py-4",
                        "pl-4 pr-4 landing-sm:pl-6 landing-sm:pr-4",
                        "rounded-[32px] landing-sm:rounded-[48px]",
                        "flex",
                        "landing-lg:hidden",
                        "items-center",
                        "justify-center",
                        "gap-2",
                        "bg-refine-blue dark:bg-refine-cyan-alt",
                        "bg-opacity-10 dark:bg-opacity-10",
                        "landing-lg:bg-opacity-100 dark:landing-lg:bg-opacity-100",
                        "text-refine-blue dark:text-refine-cyan-alt",
                        "landing-lg:text-gray-0 dark:landing-lg:text-gray-900",
                        "hover:brightness-125",
                        "landing-lg:hover:scale-105 landing-lg:hover:brightness-100",
                        "hover:text-refine-blue dark:hover:text-refine-cyan-alt",
                        "landing-lg:hover:text-gray-0 dark:landing-lg:hover:text-gray-900",
                        "landing-lg:border-8 landing-lg:border-solid",
                        activeApp.dark
                            ? "landing-lg:border-gray-900"
                            : "landing-lg:border-gray-0",
                    )}
                >
                    <span
                        className={clsx(
                            "text-xs landing-sm:text-base",
                            "font-semibold",
                        )}
                    >
                        {activeApp.label}
                    </span>
                    <LandingArrowRightIcon />
                </a>
            </div>
        </div>
    );
};


const apps = [
    {
        name: "OpenAdmin",
        link: "https://demo.openpanel.co/openadmin/",
        showcase: ShowcaseHR,
        label: "See live demo",
    },
    {
        name: "OpenPanel",
        link: "https://demo.openpanel.co/openpanel/",
        showcase: ShowcaseCRM,
        label: "See live demo",
    },
    {
        name: "OpenCLI",
        link: "https://openpanel.co/docs/category/openpanel-cli/",
        showcase: ShowcaseDevOps,
        dark: true,
        label: "Browse Commands",
    },
];
