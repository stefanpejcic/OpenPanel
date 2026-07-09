import React from "react";
import clsx from "clsx";
import { Disclosure, Transition } from "@headlessui/react";
import { CommonCircleChevronDown } from "./common-circle-chevron-down";
import { FaqSchema } from "./faq-schema";

export const EnterpriseFaq = ({ className }: { className?: string }) => {
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
        question:
            "Are there any limitations regarding the number of user accounts I can create?",
        answer: "Yes, the Community license allows you to create up to 3 user accounts.",
    },
    {
        question:
            "Are there any limitations regarding the number of domains or websites?",
        answer: "Yes, the Community license allows you to host up to 50 domains.",
    },
    {
        question:
            "Is it possible to upgrade from the Community edition to Enterprise edition?",
        answer: "Yes, at any time you can upgrade your license from Community to Enterprise edition and all limits will imediately be lifted, and additional features added.",
    },
    {
        question: "How often does Community edition receive updates?",
        answer: "The community edition receives updates on a monthly basis.",
    },
    {
        question: "Can I use OpenPanel Community for commercial hosting?",
        answer: "No, Community edition should only be used for private or personal use. For business or commercial hosting, we recommend the Enterprise edition.",
    },
    {
        question: "Is OpenPanel Community edition free forever?",
        answer: "Yes, the Community edition is free forever, with no time limit, trial period, or credit card required.",
    },
    {
        question: "Is OpenPanel open source?",
        answer: "No, OpenPanel is proprietary software, not open source. The Community edition is free to use, but the source code is not licensed for redistribution or modification.",
    },
    {
        question: "What operating systems does OpenPanel Community support?",
        answer: "OpenPanel Community currently supports Ubuntu 24.04. The Enterprise edition adds support for Debian, AlmaLinux, Rocky Linux, and CentOS.",
    },
    {
        question: "Can I migrate from cPanel or Plesk to OpenPanel?",
        answer: "Direct cPanel backup import is available in the Enterprise edition. Community users can move their sites over manually, or upgrade to Enterprise for automated migration.",
    },
];
