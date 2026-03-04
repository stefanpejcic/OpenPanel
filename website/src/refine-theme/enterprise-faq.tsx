import React from "react";
import clsx from "clsx";
import { Disclosure, Transition } from "@headlessui/react";
import Translate, { translate } from "@docusaurus/Translate";
import { CommonCircleChevronDown } from "./common-circle-chevron-down";

export const EnterpriseFaq = ({ className }: { className?: string }) => {
    return (
        <div className={clsx("flex flex-col", "not-prose", className)}>
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
                    <Translate id="enterprise.faq.title">Frequently Asked Questions</Translate>
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
                                        {translate({ message: item.question, id: item.qId })}
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
                                            {translate({ message: item.answer, id: item.aId })}
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
        qId: "enterprise.faq.q1",
        question: "How does the pricing work for the Enterprise edition?",
        aId: "enterprise.faq.a1",
        answer: "Pricing is per server, not per number of end users.",
    },
    {
        qId: "enterprise.faq.q2",
        question:
            "Are there any limitations regarding the number of user accounts I can create?",
        aId: "enterprise.faq.a2",
        answer: "Yes, the Community license allows you to create up to 3 user accounts, and Enterprise edition has no limit to the number of user accounts that you can create.",
    },
    {
        qId: "enterprise.faq.q3",
        question: "Is it possible to upgrade from the Community edition to Enterprise?",
        aId: "enterprise.faq.a3",
        answer: "Yes, you can upgrade your license from Community to Enterprise at any time, and all limitations will be immediately removed, and extra features will become available.",
    },
    {
        qId: "enterprise.faq.q4",
        question: "Do you offer a trial version for the Enterprise edition?",
        aId: "enterprise.faq.a4",
        answer: "Yes, we provide a 14-day trial period upon request.",
    },
    {
        qId: "enterprise.faq.q5",
        question: "How often are updates released for the Enterprise edition?",
        aId: "enterprise.faq.a5",
        answer: "Our goal is to provide continuous delivery of new features and fixes, potentially on a daily basis. In contrast, the Community edition receives updates on a monthly basis.",
    },
    {
        qId: "enterprise.faq.q6",
        question: "Do you offer custom development for turnkey projects?",
        aId: "enterprise.faq.a6",
        answer: "No, we do not offer any kind of turnkey development services.",
    },
    {
        qId: "enterprise.faq.q7",
        question: "What do your professional services cover?",
        aId: "enterprise.faq.a7",
        answer: "Our professional services cover collaboration with your internal teams, such as onboarding assistance, training, and customising the OpenPanel interface to match your brand.",
    },
    {
        qId: "enterprise.faq.q8",
        question: "Can I request custom features or customizations?",
        aId: "enterprise.faq.a8",
        answer: "We prioritize feature requests in our product roadmap, as well as support teams in developing custom integrations and components.",
    },
];
