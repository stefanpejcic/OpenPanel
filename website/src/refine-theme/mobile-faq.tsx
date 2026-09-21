import React from "react";
import clsx from "clsx";
import { Disclosure, Transition } from "@headlessui/react";
import { CommonCircleChevronDown } from "./common-circle-chevron-down";
import { FaqSchema } from "./faq-schema";

export const MobileFaq = ({ className }: { className?: string }) => {
    return (
        <div className={clsx("flex flex-col", "not-prose", className)}>
            <FaqSchema faq={faq} />
            <div
                className={clsx(
                    "text-2xl landing-sm:text-[32px] landing-sm:leading-[40px]",
                )}
            >
                <h2
                    className={clsx(
                        "font-semibold",
                        "dark:text-gray-400 text-gray-600",
                    )}
                >
                    Frequently Asked Questions
                </h2>
            </div>

            <div
                className={clsx(
                    "flex",
                    "flex-col",
                    "mt-6 landing-sm:mt-12 landing-lg:mt-20",
                    "not-prose",
                )}
            >
                {faq.map((item, index) => {
                    const isLast = index === faq.length - 1;

                    return (
                        <Disclosure key={index}>
                            {({ open }) => (
                                <>
                                    <Disclosure.Button
                                        className={clsx(
                                            "flex items-start justify-between",
                                            "text-start",
                                            "text-base font-semibold",
                                            "dark:text-gray-0 text-gray-900",
                                            "py-3",
                                        )}
                                    >
                                        {item.question}
                                        <CommonCircleChevronDown
                                            className={clsx(
                                                "ml-4",
                                                "flex-shrink-0",
                                                "text-gray-500",
                                                "will-change-transform",
                                                open && "transform rotate-180",
                                                "transition-transform duration-200",
                                            )}
                                        />
                                    </Disclosure.Button>
                                    <Transition
                                        unmount={false}
                                        enter="transition-all duration-300 ease-in-out"
                                        enterFrom="transform opacity-0 max-h-0"
                                        enterTo="transform opacity-100 max-h-[152px]"
                                        leave="transition-all duration-300 ease-in-out"
                                        leaveFrom="transform opacity-100 max-h-[152px]"
                                        leaveTo="transform opacity-0 max-h-0"
                                    >
                                        <Disclosure.Panel
                                            unmount={false}
                                            style={{
                                                display: "block",
                                            }}
                                            className={clsx(
                                                "mt-2 mb-6",
                                                "text-base",
                                                "dark:text-gray-400 text-gray-700",
                                            )}
                                        >
                                            {item.answer}
                                        </Disclosure.Panel>
                                    </Transition>
                                    {!isLast && (
                                        <hr
                                            className={clsx(
                                                "h-[1px]",
                                                "dark:bg-gray-700 bg-gray-200",
                                            )}
                                        />
                                    )}
                                </>
                            )}
                        </Disclosure>
                    );
                })}
            </div>
        </div>
    );
};

const faq = [
    {
        question: "Is there a limit on how many accounts or servers I can add?",
        answer: "No, add as many OpenPanel accounts and OpenAdmin servers as you manage - there's no limit.",
    },
    {
        question: "What do I need to set up on my server to connect it?",
        answer: "For OpenAdmin: API access must be enabled in its settings, and the panel's port needs to be publicly reachable over the internet. For OpenPanel: API access must be enabled for that account (check Account > API Reference - if it's not there, ask your hosting provider to enable it).",
    },
    {
        question: "Is the app free?",
        answer: "Yes, it's totally free, with no trial period, limits, or credit card required.",
    },
    {
        question: "What data do you collect about me or my servers?",
        answer: "None. No data whatsoever is recorded by us - your server details and credentials are stored only on your own device. The app's code is publicly available on GitHub, so anyone can review it.",
    },
    {
        question: "Can I back up my saved servers, or move them to another device?",
        answer: "Yes - tap the backup icon on the server list to export your saved servers and their logins to a file, which you can keep as a backup or transfer to another device and import there to restore them.",
    },
    {
        question: "Is there an iOS version?",
        answer: "No. The app is Android-only, and there's no iOS version planned.",
    },
    {
        question: "What Android version do I need?",
        answer: "Android 7.0 or newer.",
    },
    {
        question: "Do I need to create an account to use the app?",
        answer: "No - there's no separate account for the app itself. You just add your existing OpenPanel or OpenAdmin server logins.",
    },
];
