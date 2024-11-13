import clsx from "clsx";
import { useInView } from "framer-motion";
import React, {
    DetailedHTMLProps,
    FC,
    ReactNode,
    SVGProps,
    useRef,
} from "react";
import {
    Ably,
    Airtable,
    Antd,
    Appwrite,
    Chakra,
    Directus,
    Elide,
    ElideGraphql,
    Firebase,
    Graphql,
    Hasura,
    Headless,
    HookForm,
    Hygraph,
    JSONApi,
    Kbar,
    Mantine,
    Medusa,
    Mui,
    Nest,
    NestQuery,
    Nextjs,
    Remix,
    Rest,
    Sanity,
    ShadCnUI,
    SQLite,
    Strapi,
    Supabase,
    TailwindCss,
} from "../assets/integration-icons";
import { LandingSectionCtaButton } from "./landing-section-cta-button";

type Props = {
    className?: string;
};

export const LandingPackages: FC<Props> = ({ className }) => {
    return (
        <div className={clsx(className, "w-full")}>
            <div
                className={clsx("not-prose", "w-full", "px-4 landing-md:px-10")}
            >
                <h2
                    className={clsx(
                        "text-2xl landing-sm:text-[32px]",
                        "tracking-tight",
                        "text-start",
                        "p-0",
                        "dark:text-gray-0 text-gray-900",
                    )}
                >
                    Start{" "}
                    <span
                        className={clsx(
                            "font-semibold",
                            "dark:text-refine-green-alt dark:drop-shadow-[0_0_30px_rgba(71,235,235,0.3)]",
                            "text-refine-blue drop-shadow-[0_0_30px_rgba(0,128,255,0.3)]",
                        )}
                    >
                        faster
                    </span>
                    , maintain{" "}
                    <span
                        className={clsx(
                            "font-semibold",
                            "dark:text-refine-cyan-alt dark:drop-shadow-[0_0_30px_rgba(71,235,235,0.25)]",
                            "text-refine-indigo drop-shadow-[0_0_30px_rgba(51,51,255,0.3)]",
                        )}
                    >
                        easier
                    </span>
                    , scale{" "}
                    <span
                        className={clsx(
                            "font-semibold",
                            "dark:text-refine-red dark:drop-shadow-[0_0_30px_rgba(255,76,77,0.4)]",
                            "text-refine-purple drop-shadow-[0_0_30px_rgba(128,0,255,0.3)]",
                        )}
                    >
                        indefinitely
                    </span>
                    .
                </h2>
            </div>

            <div
                className={clsx(
                    "w-full",
                    "relative",
                    "mt-8 landing-sm:mt-12 landing-lg:mt-20",
                    "pb-4 landing-md:pb-10",
                    "dark:bg-landing-packages-dark bg-landing-packages",
                    "dark:bg-gray-800 bg-gray-50",
                    "rounded-2xl landing-sm:rounded-3xl",
                    "overflow-hidden",
                )}
            >
                <div
                    className={clsx(
                        "landing-packages-mask",
                        "pt-4 landing-md:pt-10",
                    )}
                >
                <PackagesContainer animDirection="right">
                    {listOne.map(({ label, tooltip }, index) => (
                        <PackageItem
                            key={index}
                            label={label}
                            tooltip={tooltip}
                        />
                    ))}
                </PackagesContainer>
                <PackagesContainer animDirection="left">
                    {listTwo.map(({ label, tooltip }, index) => (
                        <PackageItem
                            key={index}
                            label={label}
                            tooltip={tooltip}
                        />
                    ))}
                </PackagesContainer>
                </div>

                <div
                    className={clsx(
                        "not-prose",
                        "mt-4 landing-sm:mt-6 landing-lg:mt-10",
                        "px-4 landing-sm:px-10",
                    )}
                >
                    <h6
                        className={clsx(
                            "p-0",
                            "font-semibold",
                            "text-base landing-sm:text-2xl",
                            "dark:text-gray-300 text-gray-900",
                        )}
                    >
                        100+ terminal commands
                    </h6>
                    <div
                        className={clsx(
                            "not-prose",
                            "flex",
                            "items-center",
                            "justify-between",
                            "flex-wrap",
                            "gap-4 landing-sm:gap-8",
                        )}
                    >
                        <p
                            className={clsx(
                                "p-0",
                                "mt-2 landing-sm:mt-4",
                                "text-base",
                                "dark:text-gray-400 text-gray-600",
                            )}
                        >
                            OpenCLI serves as the command line interface, enabling you to execute all actions through the terminal.
                        </p>
                        <LandingSectionCtaButton to="https://dev.openpanel.com/cli/">
                            All Commands
                        </LandingSectionCtaButton>
                    </div>
                </div>
            </div>
        </div>
    );
};

const PackagesContainer = ({
    children,
    className,
    animDirection,
    ...props
}: DetailedHTMLProps<React.HTMLAttributes<HTMLDivElement>, HTMLDivElement> & {
    animDirection: "left" | "right";
}) => {
    const ref = useRef<HTMLDivElement>(null);
    const inView = useInView(ref);

    return (
        <div
            ref={ref}
            className={clsx(
                "relative",
                "flex",
                "items-center",
                animDirection === "left" ? "justify-start" : "justify-end",
            )}
        >
            <div
                className={clsx(
                    className,
                    "hover:animation-paused",
                    inView
                        ? animDirection === "left"
                            ? "animate-landing-packages-left"
                            : "animate-landing-packages-right"
                        : "",
                    "absolute",
                    "left-0",
                    "top-0",
                    "pr-4",
                    "w-auto",
                    "flex",
                    "items-center",
                    "gap-[18px]",
                    "mt-6",
                    "relative",
                )}
                {...props}
            >
                {children}
            </div>
        </div>
    );
};

const PackageItem = (props: {
    icon: ReactNode;
    label: string;
    tooltip: string;
}) => {
    const { tooltip, icon, label } = props;

    return (
        <div
            className={clsx(
                "group",
                "relative",
                "z-10",
                "flex",
                "items-center",
                "justify-center",
                "gap-3",
                "pl-4 pt-4 pb-4 pr-6",
                "dark:bg-gray-900 bg-gray-0",
                "rounded-full",
                "cursor-pointer",
            )}
        >
            <div>{icon}</div>
            <div
                className={clsx(
                    "text-sm",
                    "font-medium",
                    "dark:bg-landing-packages-text-dark bg-landing-packages-text",
                    "bg-clip-text",
                    "text-transparent",
                    "whitespace-nowrap",
                )}
            >
                {label}
            </div>

            <div
                className={clsx(
                    "absolute",
                    "z-20",
                    "top-[-48px]",
                    "scale-0",
                    "group-hover:scale-100",
                    "transition-transform",
                    "origin-top",
                )}
            >
                <div
                    className={clsx(
                        "relative",
                        "text-sm",
                        "dark:bg-gray-0 bg-gray-900",
                        "dark:text-gray-700 text-gray-300",
                        "rounded-full",
                        "px-6",
                        "py-3",
                        "whitespace-nowrap",
                    )}
                >
                    {tooltip}
                </div>

                <svg
                    xmlns="http://www.w3.org/2000/svg"
                    width={40}
                    height={15}
                    fill="none"
                    className={clsx(
                        "absolute",
                        "scale-0",
                        "-bottom-2",
                        "left-1/2",
                        "-translate-x-1/2",
                        "group-hover:scale-100",
                        "transition-transform",
                        "origin-bottom",
                        "dark:text-gray-0 text-gray-900",
                    )}
                >
                    <path
                        fill="currentColor"
                        d="M17.73 13.664C18.238 14.5 19.089 15 20 15c.912 0 1.763-.501 2.27-1.336l3.025-4.992C26.306 7.002 28.01 7 29.833 7H40V0H0v7h10.167c1.823 0 3.527.003 4.538 1.672l3.026 4.992Z"
                    />
                </svg>
            </div>
        </div>
    );
};

const listOne = [
    {
        label: "Disable admin panel",
        tooltip: "opencli admin off",
    },
    {
        label: "List Users",
        tooltip: "opencli user-list",
    },
    {
        label: "Reset Password",
        tooltip: "opencli user-password <USERNAME> <NEW_PASSWORD>",
    },
    {
        label: "Set nameservers",
        tooltip: "opencli config update ns1 <NAMESERVER>",
    },
    {
        label: "View logout URL",
        tooltip: "opencli config get logout_url",
    },
    {
        label: "List plans",
        tooltip: "opencli plan-list",
    },
    {
        label: "Disable 2FA",
        tooltip: "opencli user-2fa <USERNAME> disable",
    },
    {
        label: "View Backup Jobs",
        tooltip: "opencli backup-job list",
    },
    {
        label: "List SSLs",
        tooltip: "opencli ssl-user <USERNAME>",
    },
    {
        label: "Add a new user",
        tooltip: "opencli user-add <USERNAME> <PASSWORD> <EMAIL> <PLAN_ID>",
    },
    {
        label: "List backups",
        tooltip: "opencli backup-list <USERNAME>",
    },
    {
        label: "Set brand name",
        tooltip: "opencli config update brand_name <NAME>",
    },
    {
        label: "Change PHP version for domain",
        tooltip: "opencli php-domain_php <DOMAIN-NAME> --update <PHP-VERSION>",
    },
    {
        label: "Who owns Domain",
        tooltip: "opencli domains-whoowns <DOMAIN-NAME>",
    },
    {
        label: "Assign IP",
        tooltip: "opencli user-ip <USERNAME> <IP_ADDRESS>",
    },
    {
        label: "Enable Memcached",
        tooltip: "opencli user-memcached enable <USERNAME>",
    },
];

const listTwo = [
    {
        label: "Change OpenPanel port",
        tooltip: "opencli config update port <NEW-PORT>",
    },
    {
        label: "Add Backup Destination",
        tooltip: "opencli backup-destination create <HOSTNAME> <PASSWORD> <PORT> <USER> <PATH_TO_SSH_KEY_FILE> <STORAGE_PERCENTAGE>",
    },
    {
        label: "Generate hostname SSL",
        tooltip: "opencli ssl-hostname",
    },
    {
        label: "View Login Log",
        tooltip: "opencli user-loginlog <USERNAME>",
    },
    {
        label: "Install new PHP version",
        tooltip: "opencli php-install_php_version <USERNAME> <PHP-VERSION>",
    },
    {
        label: "List user domains",
        tooltip: "opencli domains-user <USERNAME>",
    },
    {
        label: "Suspend user",
        tooltip: "opencli user-suspend <USERNAME>",
    },
    {
        label: "Re-index backups",
        tooltip: "opencli backup-index <ID>",
    },
    {
        label: "Get PHP version for domain",
        tooltip: "opencli php-domain_php <DOMAIN-NAME>",
    },
    {
        label: "Check Resource Usage",
        tooltip: "opencli docker-collect_stats",
    },
    {
        label: "Login to User",
        tooltip: "opencli user-login <USERNAME>",
    },
    {
        label: "Restore from backup",
        tooltip: "opencli backup-restore <DATE> <USER> --all",
    },
    {
        label: "Fix Permissions",
        tooltip: "opencli files-fix_permissions [USERNAME] [PATH]",
    },
    {
        label: "Install ModSecurity",
        tooltip: "opencli nginx-install_modsec",
    },
];
