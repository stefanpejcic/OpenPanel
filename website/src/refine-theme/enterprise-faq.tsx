import React from "react";
import clsx from "clsx";
import { Disclosure, Transition } from "@headlessui/react";
import { CommonCircleChevronDown } from "./common-circle-chevron-down";
import { FaqSchema } from "./faq-schema";

export const EnterpriseFaq = ({ className }: { className?: string }) => {
    return (
        <div className={clsx("flex flex-col", "not-prose", className)}>
            <FaqSchema
                faq={faq.map((item) => ({
                    question: item.question,
                    answer:
                        typeof item.answer === "string"
                            ? item.answer
                            : item.schemaAnswer,
                }))}
            />
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
        question: "How does the pricing work for the Enterprise edition?",
        answer: "Pricing is per server, not per number of end users.",
    },
    {
        question:
            "Are there any limitations regarding the number of user accounts I can create?",
        answer: "Yes, the Community license allows you to create up to 3 user accounts, and Enterprise edition has no limit to the number of user accounts that you can create.",
    },
    {
        question:
            "Is it possible to upgrade from the Community edition to Enterprise edition?",
        answer: "Yes, at any time you can upgrade your license from Community to Enterprise edition and all limits will imediately be lifted, and additional features added.",
    },
    {
        question:
            "Do you offer trial license for the Enterprise edition?",
        answer: "Yes — click \"Start free trial\" above to get instant access to a 30-day Enterprise trial, no credit card required.",
    },
    {
        question: "What happens after my trial ends?",
        answer: "If you don't purchase before the trial ends, the OpenAdmin interface is blocked until you remove the license key - your websites and the OpenPanel user interface keep working normally the whole time. Once you remove the key, everything reverts to Community edition limits. No automatic charge, no surprise billing.",
    },
    {
        question: "How often does Enterprise edition receive updates?",
        answer: "We aim to introduce new features and fixes in a continuous delivery manner, sometimes as frequently as daily updates. On the other hand, the community edition receives updates on a monthly basis.",
    },
    {
        question: "What operating systems does OpenPanel Enterprise support?",
        answer: "OpenPanel Enterprise supports Ubuntu, Debian, AlmaLinux, Rocky Linux, and CentOS.",
    },
    {
        question: "Does OpenPanel run on ARM (AArch64) servers?",
        answer: "Yes, OpenPanel Enterprise runs on both standard x86_64 servers and ARM (AArch64) CPUs.",
    },
    {
        question: "Can I migrate from cPanel to OpenPanel?",
        answer: "Yes, OpenPanel Enterprise can import accounts directly from a cPanel backup.",
    },
    {
        question: "Does OpenPanel Enterprise integrate with billing software?",
        answer: "Yes, OpenPanel Enterprise integrates with WHMCS, Blesta, FOSSBilling, ClientExec, WISECP, and Paymenter.org for automated provisioning.",
    },
    {
        question: "What IP is required to purchase a license?",
        answer: (
            <>
                As OpenPanel does not provide web hosting services,
                purchasing a monthly OpenPanel Enterprise license requires
                the public-facing static IPv4 address of the server. The
                server must also meet the minimum requirements listed in the{" "}
                <a href="/docs/admin/intro/#requirements">
                    System Requirements documentation
                </a>
                .
            </>
        ),
        schemaAnswer:
            "As OpenPanel does not provide web hosting services, purchasing a monthly OpenPanel Enterprise license requires the public-facing static IPv4 address of the server. The server must also meet the minimum requirements listed in the System Requirements documentation.",
    },
    {
        question: "What payment methods does OpenPanel accept?",
        answer: "The OpenPanel store accepts Credit Card and Stripe payment methods. You can add and manage payment methods by logging into your store account and going to Account Settings, then Payment Methods.",
    },
    {
        question: "When does my OpenPanel license renew?",
        answer: (
            <>
                All OpenPanel licenses renew automatically on the same day of
                the month (or year, for annual licenses) on which the
                service was originally purchased. At renewal, the primary
                payment method on your account is charged for your licenses.
                To prevent a license from renewing, cancel it before the
                renewal date — see{" "}
                <a href="/docs/articles/license/hot_to_cancel_openpanel_license/">
                    How do I cancel my OpenPanel Enterprise license?
                </a>
            </>
        ),
        schemaAnswer:
            "All OpenPanel licenses renew automatically on the same day of the month (or year, for annual licenses) on which the service was originally purchased. At renewal, the primary payment method on your account is charged for your licenses. To prevent a license from renewing, cancel it before the renewal date - see How do I cancel my OpenPanel Enterprise license?",
    },
    {
        question: "Do additional IP addresses need to be licensed?",
        answer: "No, additional IP addresses do not need to be licensed - only the server's main, public IPv4 address is required to be licensed. Additional IP addresses can be used for other purposes, such as being dedicated to an account, used for certain services, or reserved for later use.",
    },
    {
        question: "Is there a refund if I'm not satisfied?",
        answer: "Refunds are available within 7 days of purchase, but only for licenses that haven't been activated yet. Once a license is activated on a server, the sale is final.",
    },
    {
        question: "What professional services do you provide?",
        answer: "We offer onboarding assistance, training, and customization of your OpenPanel instance to match your brand, and we prioritize feature requests from Enterprise customers on our product roadmap.",
    },
];
