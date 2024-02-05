import clsx from "clsx";
import React, { FC } from "react";
import { CommonCircleChevronDown } from "./common-circle-chevron-down";
import { CommonCircleChevronUp } from "./common-circle-chevron-up";
import { LandingSectionCtaButton } from "./landing-section-cta-button";

type Props = {
    className?: string;
};

export const LandingTestimonial: FC<Props> = ({ className }) => {
    const [showMore, setShowMore] = React.useState(false);

    const testimonialsFirstThree = testimonials.slice(0, 3);
    const testiomianlsStartFromThree = testimonials.slice(3);
    const testimonialsFirstTwo = testimonials.slice(0, 2);
    const testimonialsStartFromTwo = testimonials.slice(2);

    return (
        <div
            className={clsx(
                className,
                "w-full",
                "flex",
                "flex-col",
                "items-center",
            )}
        >
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
                    The{" "}
                    <span
                        className={clsx(
                            "font-semibold",
                            "dark:text-refine-cyan-alt dark:drop-shadow-[0_0_30px_rgba(71,235,235,0.25)]",
                            "text-refine-blue drop-shadow-[0_0_30px_rgba(0,128,255,0.3)]",
                        )}
                    >
                        difference
                    </span>{" "}
                    that OpenPanel makes
                </h2>
            </div>

            <div
                className={clsx(
                    "mt-8 landing-sm:mt-12 landing-lg:mt-20",
                    "grid grid-cols-12",
                    "gap-6",
                    "items-stretch",
                )}
            >
                {testimonialsFirstThree.map((testimonial, index) => {
                    return (
                        <TestimonialCard
                            key={index}
                            className={clsx(
                                "block landing-md:hidden landing-lg:block",
                                "col-span-full landing-lg:col-span-4",
                                "h-auto",
                            )}
                            {...testimonial}
                        />
                    );
                })}
                {testimonialsFirstTwo.map((testimonial, index) => {
                    return (
                        <TestimonialCard
                            key={index}
                            className={clsx(
                                "hidden landing-md:block landing-lg:hidden",
                                "col-span-6",
                                "h-auto",
                            )}
                            {...testimonial}
                        />
                    );
                })}
            </div>
            {showMore && (
                <div
                    className={clsx(
                        "block landing-md:hidden landing-lg:block",
                        "columns-1 landing-lg:columns-3 gap-6",
                    )}
                >
                    {testiomianlsStartFromThree.map((testimonial, index) => {
                        return (
                            <TestimonialCard
                                key={index}
                                className={clsx(
                                    "inline-block",
                                    "w-full",
                                    "mt-6",
                                )}
                                {...testimonial}
                            />
                        );
                    })}
                </div>
            )}
            {showMore && (
                <div
                    className={clsx(
                        "hidden landing-md:block landing-lg:hidden",
                        "columns-2 gap-6",
                    )}
                >
                    {testimonialsStartFromTwo.map((testimonial, index) => {
                        return (
                            <TestimonialCard
                                key={index}
                                className={clsx(
                                    "inline-block",
                                    "w-full",
                                    "mt-6",
                                )}
                                {...testimonial}
                            />
                        );
                    })}
                </div>
            )}
            <LandingSectionCtaButton
                className={clsx("cursor-pointer", "mt-6")}
                onClick={() => setShowMore((prev) => !prev)}
                icon={
                    showMore ? (
                        <CommonCircleChevronUp />
                    ) : (
                        <CommonCircleChevronDown />
                    )
                }
            >
                Show {showMore ? "less" : "more"}
            </LandingSectionCtaButton>
        </div>
    );
};

const TestimonialCard = ({
    name,
    title,
    description,
    img,
    link,
    className,
}) => {
    return (
        <div
            className={clsx(
                "border dark:border-gray-700 border-gray-200",
                "rounded-3xl",
                className,
            )}
        >
            <div
                className={clsx(
                    "not-prose",
                    "h-full",
                    "flex flex-col",
                    "justify-between",
                    "gap-6",
                    "p-10",
                )}
            >
                <div
                    className={clsx(
                        "text-base",
                        "dark:text-gray-0 text-gray-900",
                    )}
                >
                    {description}
                </div>
                <a
                    href={link}
                    target="_blank"
                    rel="noopener noreferrer"
                    className={clsx(
                        "flex gap-4",
                        "items-center",
                        "appearance-none",
                        "no-underline",
                    )}
                >
                    <img
                        src={img}
                        alt={name}
                        className={clsx("w-12 h-12 rounded-full")}
                    />
                    <div className={clsx("flex flex-col")}>
                        <div
                            className={clsx(
                                "text-sm",
                                "dark:text-gray-400 text-gray-600",
                            )}
                        >
                            {name}
                        </div>
                        <div
                            className={clsx(
                                "text-sm",
                                "dark:text-gray-400 text-gray-600",
                            )}
                        >
                            {title}
                        </div>
                    </div>
                </a>
            </div>
        </div>
    );
};

const testimonials = [
    {
        title: "CEO - Ja doo",
        name: "ZJohn Doe",
        link: "https://google.com",
        description:
            "nestoooo",
        img: "https://refine.ams3.cdn.digitaloceanspaces.com/website/static/assets/testimonials/zeno-rocha.png",
    },
    {
        title: "CEO - Ja doo",
        name: "ZJohn Doe",
        link: "https://google.com",
        description:
            "nestoooo",
        img: "https://refine.ams3.cdn.digitaloceanspaces.com/website/static/assets/testimonials/zeno-rocha.png",
    },
    {
        title: "CEO - Ja doo",
        name: "ZJohn Doe",
        link: "https://google.com",
        description:
            "nestoooo",
        img: "https://refine.ams3.cdn.digitaloceanspaces.com/website/static/assets/testimonials/zeno-rocha.png",
    },
    {
        title: "CEO - Ja doo",
        name: "ZJohn Doe",
        link: "https://google.com",
        description:
            "nestoooo",
        img: "https://refine.ams3.cdn.digitaloceanspaces.com/website/static/assets/testimonials/zeno-rocha.png",
    },
    {
        title: "CEO - Ja doo",
        name: "ZJohn Doe",
        link: "https://google.com",
        description:
            "nestoooo",
        img: "https://refine.ams3.cdn.digitaloceanspaces.com/website/static/assets/testimonials/zeno-rocha.png",
    },
    {
        title: "CEO - Ja doo",
        name: "ZJohn Doe",
        link: "https://google.com",
        description:
            "nestoooo",
        img: "https://refine.ams3.cdn.digitaloceanspaces.com/website/static/assets/testimonials/zeno-rocha.png",
    },
    {
        title: "CEO - Ja doo",
        name: "ZJohn Doe",
        link: "https://google.com",
        description:
            "nestoooo",
        img: "https://refine.ams3.cdn.digitaloceanspaces.com/website/static/assets/testimonials/zeno-rocha.png",
    },
    {
        title: "CEO - Ja doo",
        name: "ZJohn Doe",
        link: "https://google.com",
        description:
            "nestoooo",
        img: "https://refine.ams3.cdn.digitaloceanspaces.com/website/static/assets/testimonials/zeno-rocha.png",
    },
    {
        title: "CEO - Ja doo",
        name: "ZJohn Doe",
        link: "https://google.com",
        description:
            "nestoooo",
        img: "https://refine.ams3.cdn.digitaloceanspaces.com/website/static/assets/testimonials/zeno-rocha.png",
    },
    {
        title: "CEO - Ja doo",
        name: "ZJohn Doe",
        link: "https://google.com",
        description:
            "nestoooo",
        img: "https://refine.ams3.cdn.digitaloceanspaces.com/website/static/assets/testimonials/zeno-rocha.png",
    },
];
