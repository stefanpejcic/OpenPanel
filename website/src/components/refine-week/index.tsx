import React from "react";
import clsx from "clsx";
import { data as weekData } from "./data";
import { RefineWeekDesktop } from "./refine-week-desktop";
import { RefineWeekMobile } from "./refine-week-mobile";

type Props = {
    variant: "strapi" | "supabase";
};

export const RefineWeek = ({ variant }: Props) => {
    return (
        <div className={clsx("pb-20", "overflow-hidden")}>
            <RefineWeekDesktop
                className={clsx("hidden lg:block")}
                variant={variant}
                data={weekData[variant]}
            />
            <RefineWeekMobile
                className={clsx("block lg:hidden")}
                variant={variant}
                data={weekData[variant]}
            />

            <div
                className={clsx(
                    "max-w-[620px] lg:max-w-[944px]",
                    "mx-auto mt-20",
                )}
            >
                <div
                    className={clsx(
                        "max-w-[344px] sm:max-w-[620px] lg:max-w-[944px] mx-auto",
                        "text-start lg:text-center",
                        "text-[32px] leading-[40px] font-bold",
                        "text-gray-700 dark:text-gray-300",
                    )}
                >
                    Supported Billing Integrations
                </div>

                <div
                    className={clsx(
                        "flex flex-wrap gap-4 justify-center items-center",
                        "mt-6 lg:mt-10",
                    )}
                >
                    {[
                        {
                            title: "FOSSBilling",
                            label: "openpanel.com",
                            link: "https://openpanel.com/docs/articles/extensions/openpanel-and-fossbilling/",
                        },
                        {
                            title: "WHMCS",
                            label: "openpanel.com",
                            link: "https://openpanel.com/docs/articles/extensions/openpanel-and-whmcs/",
                        },
                        {
                            title: "Blesta",
                            label: "openpanel.com",
                            link: "https://openpanel.com/docs/articles/extensions/openpanel-and-blesta/",
                        },
                    ].map((item) => (
                        <a
                            key={item.title}
                            href={item.link}
                            target="_blank"
                            rel="noreferrer"
                            className={clsx(
                                "no-underline",
                                "flex flex-col gap-1 justify-start",
                                "w-[160px] sm:w-[192px]",
                                "p-4",
                                "rounded-2xl",
                                "border border-gray-200 dark:border-gray-700",
                                "text-xs sm:text-base",
                            )}
                        >
                            <div className="text-gray-500 dark:text-gray-400">
                                {item.title}
                            </div>
                            <div className="text-gray-900 dark:text-gray-0 font-semibold">
                                {item.label}
                            </div>
                        </a>
                    ))}
                </div>
            </div>
        </div>
    );
};
