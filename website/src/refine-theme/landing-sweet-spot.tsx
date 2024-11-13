import React, { FC, useEffect, useRef, useState } from "react";
import clsx from "clsx";
import { useColorMode } from "@docusaurus/theme-common";
import {
    AuthenticationIcon,
    ChartsIcon,
    DataTablesIcon,
    FormsIcon,
    ListIcon,
    WizardsIcon,
} from "../components/landing/icons";
import useIsBrowser from "@docusaurus/useIsBrowser";
import { AnimatePresence, motion, useInView } from "framer-motion";

type Props = {
    className?: string;
};

export const LandingSweetSpot: FC<Props> = ({ className }) => {
    const ref = useRef<HTMLDivElement>(null);
    const inView = useInView(ref);

    const isBrowser = useIsBrowser();

    const { colorMode } = useColorMode();
    const isDarkTheme = colorMode === "dark";

    const [activeIndex, setActiveIndex] = useState(0);
    const activeListItem = list[activeIndex];

    const [shouldIncrement, setShouldIncrement] = useState(true);

    useEffect(() => {
        if (!shouldIncrement) {
            return;
        }

        let interval: NodeJS.Timeout;
        if (inView) {
            interval = setInterval(() => {
                setActiveIndex((prev) => (prev + 1) % list.length);
            }, 3000);
        }

        return () => clearInterval(interval);
    }, [shouldIncrement, inView]);

    return (
        <div ref={ref} className={clsx(className, "w-full")}>
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
                            "text-refine-indigo drop-shadow-[0_0_30px_rgba(51,51,255,0.3)]",
                        )}
                    >
                        sweet spot
                    </span>{" "}
                    between shared hosting and VPS.
                </h2>
                <p
                    className={clsx(
                        "mt-4 landing-sm:mt-6",
                        "max-w-md",
                        "text-base",
                        "dark:text-gray-400 text-gray-600",
                    )}
                >
                    Every OpenPanel user enjoys a dedicated and isolated environment where they can install services, define PHP limits, install new versions, and configure various components such as Redis and Elasticsearch.
                </p>
            </div>

            <div className={clsx("mt-8 landing-sm:mt-12 landing-lg:mt-20")}>
                <div
                    className={clsx(
                        "select-none",
                        "relative",
                        "h-[752px] landing-sm:h-[874px] landing-md:h-[984px] landing-lg:h-[688px]",
                        "not-prose",
                        "pt-4 landing-sm:pt-10 landing-lg:pt-20",
                        "pb-4 landing-lg:pb-0",
                        "pl-4 landing-sm:pl-10",
                        "dark:bg-gray-800 bg-gray-50",
                        "rounded-2xl landing-sm:rounded-3xl",
                        "overflow-hidden",
                        "dark:bg-noise",
                    )}
                >
                    <AnimatePresence>
                        <motion.div
                            initial={{ opacity: 0 }}
                            animate={{ opacity: 1 }}
                            exit={{ opacity: 0 }}
                            transition={{
                                duration: 0.5,
                                ease: "easeInOut",
                            }}
                            key={activeIndex}
                            className={clsx(
                                "absolute",
                                "inset-0",
                                "z-0",
                                "landing-xs:bg-landing-sweet-spot-glow-position-xs",
                                "landing-lg:bg-landing-sweet-spot-glow-position-lg",
                                "landing-md:bg-landing-sweet-spot-glow-position-md",
                                "landing-xs:bg-landing-sweet-spot-glow-size-xs",
                                "landing-lg:bg-landing-sweet-spot-glow-size-lg",
                                activeListItem.backgroundImage,
                            )}
                            style={{
                                backgroundRepeat: "repeat, no-repeat",
                            }}
                        />
                    </AnimatePresence>
                    <div
                        className={clsx(
                            "relative",
                            "z-[1]",
                            "h-full w-full",
                            "flex flex-col landing-lg:grid landing-lg:grid-cols-12",
                        )}
                    >
                        <div
                            className={clsx(
                                "not-prose",
                                "pr-6 landing-sm:pr-0",
                                "landing-sm:max-w-[540px] landing-md:max-w-[760px] landing-lg:max-w-[416px]",
                                "landing-lg:col-span-5",
                                "landing-lg:mt-16",
                            )}
                        >
                            <h3
                                className={clsx(
                                    "text-base landing-sm:text-xl font-semibold",
                                    "dark:text-gray-300 text-gray-700",
                                )}
                            >
                                {activeListItem.title}
                            </h3>
                            <p
                                className={clsx(
                                    "mt-6",
                                    "text-base",
                                    "dark:text-gray-400 text-gray-600",
                                )}
                            >
                                {activeListItem.description}
                            </p>
                            <div
                                className={clsx(
                                    "mt-4 landing-sm:mt-10",
                                    "w-max",
                                    "grid",
                                    "grid-cols-2 landing-sm:grid-cols-3 landing-lg:grid-cols-2",
                                    "landing-sm:gap-x-2 gap-y-4",
                                    "not-prose",
                                )}
                            >
                                {list.map((item, index) => {
                                    const active = index === activeIndex;
                                    const Icon = item.icon;

                                    return (
                                        <button
                                            key={item.iconText}
                                            onClick={() => {
                                                setShouldIncrement(false);
                                                setActiveIndex(index);
                                            }}
                                            className={clsx(
                                                "appearance-none",
                                                "focus:outline-none",
                                                "cursor-pointer",
                                                active
                                                    ? "dark:bg-gray-900 bg-gray-0"
                                                    : "dark:bg-gray-900/50 bg-gray-0/50",

                                                "w-max",
                                                "flex",
                                                "items-center",
                                                "justify-start",
                                                "gap-1",
                                                "px-4 py-2",
                                                "rounded-full",
                                                "text-sm landing-sm:text-base",
                                            )}
                                        >
                                            <div>
                                                <Icon active={active} />
                                            </div>
                                            <div
                                                className={clsx(
                                                    active
                                                        ? "dark:text-gray-0 text-gray-900"
                                                        : "dark:text-gray-400 text-gray-600",
                                                )}
                                            >
                                                {item.iconText}
                                            </div>
                                        </button>
                                    );
                                })}
                            </div>
                        </div>
                        {isBrowser && (
                            <div
                                className={clsx(
                                    "relative",
                                    "h-full",
                                    "mt-4 landing-sm:mt-[72px] landing-lg:mt-0",
                                    "flex",
                                    "landing-lg:col-start-7 landing-lg:col-end-13",
                                )}
                            >
                                <div
                                    className={clsx(
                                        "w-full",
                                        "h-full",
                                        "landing-sweet-spot-mask",
                                        "z-[1]",
                                        "landing-lg:absolute",
                                        "top-0 right-0",
                                    )}
                                >
                                    {list.map((item, index) => {
                                        const active = index === activeIndex;

                                        return (
                                            <img
                                                key={index}
                                                src={
                                                    isDarkTheme
                                                        ? item.image1Dark
                                                        : item.image1Light
                                                }
                                                alt="UI of refine"
                                                className={clsx(
                                                    "block",
                                                    "object-cover",
                                                    "object-left-top",
                                                    "w-full landing-md:w-[874px] landing-lg:w-full",
                                                    "h-full landing-lg:h-[464px]",
                                                    "landing-md:pl-20 landing-lg:pl-0",
                                                    "absolute",
                                                    "top-0 right-0",
                                                    active && "delay-300",
                                                    active
                                                        ? "translate-x-0"
                                                        : "translate-x-full",
                                                    active
                                                        ? "opacity-100"
                                                        : "opacity-0",
                                                    "transition-[transform,opacity] duration-500 ease-in-out",
                                                )}
                                            />
                                        );
                                    })}
                                </div>

                                {list.map((item, index) => {
                                    const active = index === activeIndex;

                                    return (
                                        <div
                                            key={index}
                                            className={clsx(
                                                "block",
                                                "z-[2]",
                                                "w-[328px] landing-sm:w-[488px]",
                                                "absolute",
                                                "bottom-0 landing-sm:bottom-[4px] landing-lg:bottom-[78px]",
                                                "-left-2 landing-lg:-left-20",
                                                "rounded-xl",
                                                "dark:bg-gray-900 bg-gray-0",
                                                "dark:shadow-landing-sweet-spot-code-dark",
                                                "shadow-landing-sweet-spot-code-light",
                                                active && "delay-300",
                                                active
                                                    ? "translate-y-0"
                                                    : "translate-y-full",
                                                active
                                                    ? "opacity-100"
                                                    : "opacity-0",
                                                "transition-[transform,opacity] duration-500 ease-in-out",
                                            )}
                                        >
                                            <img
                                                src={
                                                    isDarkTheme
                                                        ? item.image2Dark
                                                        : item.image2Light
                                                }
                                                alt="Code of OpenPanel"
                                            />
                                        </div>
                                    );
                                })}
                            </div>
                        )}
                    </div>
                </div>
            </div>
        </div>
    );
};

const list = [
    {
        title: "Nginx or Apache web server",
        description: `For each user you can choose to use Nginx or Apache web server.`,
        icon: (props: { active: boolean }) => (
        <svg width="24" height="24" viewBox="-17.5 0 291 291" version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" preserveAspectRatio="xMidYMid">
            <g>
                        <path d="M2.05386819,218.186819 C3.37421203,220.534097 5.28137536,222.294556 7.6286533,223.6149 L120.591404,288.751862 L120.591404,288.751862 C125.28596,291.539255 131.00745,291.539255 135.555301,288.751862 L248.518052,223.6149 C253.212607,220.974212 256,215.986246 256,210.558166 L256,80.2842407 L256,80.2842407 C256,74.8561605 253.212607,69.8681948 248.518052,67.2275072 L135.555301,2.09054441 L135.555301,2.09054441 C130.860745,-0.696848138 125.139255,-0.696848138 120.591404,2.09054441 L120.591404,2.09054441 L7.6286533,67.2275072 C2.78739255,69.8681948 0,74.8561605 0,80.2842407 L0,80.2842407 L0,210.704871 C0,213.345559 0.586819484,215.839542 2.05386819,218.186819" fill="#009639">

        </path>
                        <path d="M91.8372493,195.154155 C91.8372493,203.222923 85.382235,209.677937 77.313467,209.677937 C69.2446991,209.677937 62.7896848,203.222923 62.7896848,195.154155 L62.7896848,195.154155 L62.7896848,95.5415473 C62.7896848,87.7661891 69.6848138,81.4578797 79.2206304,81.4578797 C86.1157593,81.4578797 94.1845272,84.2452722 99.025788,90.2601719 L103.426934,95.5415473 L164.162751,168.160458 L164.162751,95.834957 L164.162751,95.834957 C164.162751,87.7661891 170.617765,81.3111748 178.686533,81.3111748 C186.755301,81.3111748 193.210315,87.7661891 193.210315,95.834957 L193.210315,95.834957 L193.210315,195.447564 C193.210315,203.222923 186.315186,209.531232 176.77937,209.531232 C169.884241,209.531232 161.815473,206.74384 156.974212,200.72894 L91.8372493,122.975358 L91.8372493,195.154155 L91.8372493,195.154155 Z" fill="#FFFFFF">

        </path>
            </g>
        </svg>
        ),
        iconText: "Nginx",
        image1Dark:
            "/img/ilustrations/nginx_large.png",
        image1Light:
            "/img/ilustrations/nginx_large.png",
        image2Dark:
            "/img/ilustrations/nginx_small.png",
        image2Light:
            "/img/ilustrations/nginx_small.png",
        backgroundImage:
            "dark:bg-landing-sweet-spot-glow-red-dark bg-landing-sweet-spot-glow-red-light",
    },
    {
        title: "Isolated user accounts with Docker",
        description: `Each user account is an isolated Docker container where user can install services and edit configuration.`,
        icon: (props: { active: boolean }) => (
        <svg width="24" height="24" viewBox="0 0 16 16" xmlns="http://www.w3.org/2000/svg" fill="none">
            <path
                fill="#2396ED"
                d="M12.342 4.536l.15-.227.262.159.116.083c.28.216.869.768.996 1.684.223-.04.448-.06.673-.06.534 0 .893.124 1.097.227l.105.057.068.045.191.156-.066.2a2.044 2.044 0 01-.47.73c-.29.299-.8.652-1.609.698l-.178.005h-.148c-.37.977-.867 2.078-1.702 3.066a7.081 7.081 0 01-1.74 1.488 7.941 7.941 0 01-2.549.968c-.644.125-1.298.187-1.953.185-1.45 0-2.73-.288-3.517-.792-.703-.449-1.243-1.182-1.606-2.177a8.25 8.25 0 01-.461-2.83.516.516 0 01.432-.516l.068-.005h10.54l.092-.007.149-.016c.256-.034.646-.11.92-.27-.328-.543-.421-1.178-.268-1.854a3.3 3.3 0 01.3-.81l.108-.187zM2.89 5.784l.04.007a.127.127 0 01.077.082l.006.04v1.315l-.006.041a.127.127 0 01-.078.082l-.039.006H1.478a.124.124 0 01-.117-.088l-.007-.04V5.912l.007-.04a.127.127 0 01.078-.083l.039-.006H2.89zm1.947 0l.039.007a.127.127 0 01.078.082l.006.04v1.315l-.007.041a.127.127 0 01-.078.082l-.039.006H3.424a.125.125 0 01-.117-.088L3.3 7.23V5.913a.13.13 0 01.085-.123l.039-.007h1.413zm1.976 0l.039.007a.127.127 0 01.077.082l.007.04v1.315l-.007.041a.127.127 0 01-.078.082l-.039.006H5.4a.124.124 0 01-.117-.088l-.006-.04V5.912l.006-.04a.127.127 0 01.078-.083l.039-.006h1.413zm1.952 0l.039.007a.127.127 0 01.078.082l.007.04v1.315a.13.13 0 01-.085.123l-.04.006H7.353a.124.124 0 01-.117-.088l-.006-.04V5.912l.006-.04a.127.127 0 01.078-.083l.04-.006h1.412zm1.97 0l.039.007a.127.127 0 01.078.082l.006.04v1.315a.13.13 0 01-.085.123l-.039.006H9.322a.124.124 0 01-.117-.088l-.006-.04V5.912l.006-.04a.127.127 0 01.078-.083l.04-.006h1.411zM4.835 3.892l.04.007a.127.127 0 01.077.081l.007.041v1.315a.13.13 0 01-.085.123l-.039.007H3.424a.125.125 0 01-.117-.09l-.007-.04V4.021a.13.13 0 01.085-.122l.039-.007h1.412zm1.976 0l.04.007a.127.127 0 01.077.081l.007.041v1.315a.13.13 0 01-.085.123l-.039.007H5.4a.125.125 0 01-.117-.09l-.006-.04V4.021l.006-.04a.127.127 0 01.078-.082l.039-.007h1.412zm1.953 0c.054 0 .1.037.117.088l.007.041v1.315a.13.13 0 01-.085.123l-.04.007H7.353a.125.125 0 01-.117-.09l-.006-.04V4.021l.006-.04a.127.127 0 01.078-.082l.04-.007h1.412zm0-1.892c.054 0 .1.037.117.088l.007.04v1.316a.13.13 0 01-.085.123l-.04.006H7.353a.124.124 0 01-.117-.088l-.006-.04V2.128l.006-.04a.127.127 0 01.078-.082L7.353 2h1.412z"
            />
        </svg>
        ),
        iconText: "Docker",
        image1Dark:
            "/img/ilustrations/docker_large.png",
        image1Light:
            "/img/ilustrations/docker_large.png",
        image2Dark:
            "/img/ilustrations/docker_small.png",
        image2Light:
            "/img/ilustrations/docker_small.png",
        backgroundImage:
            "dark:bg-landing-sweet-spot-glow-orange-dark bg-landing-sweet-spot-glow-orange-light",
    },
    {
        title: "Separate MySQL instances",
        description: `Each user has their own MySQL or MariaDB instance and can edit configuration settings for their databases.`,
        icon: (props: { active: boolean }) => (
        <svg width="24" height="24" viewBox="0 -2 256 256" version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" preserveAspectRatio="xMidYMid">
            <path d="M235.648276,194.211632 C221.729851,193.864559 210.942872,195.257272 201.895604,199.083964 C199.285899,200.127406 195.109927,200.128498 194.761767,203.433458 C196.154498,204.826189 196.328034,207.087716 197.546096,209.001055 C199.635192,212.479551 203.287247,217.178343 206.593317,219.614507 C210.246461,222.397748 213.900691,225.180989 217.727416,227.617153 C224.513092,231.793125 232.168625,234.22817 238.779677,238.404123 C242.608577,240.838067 246.434123,243.971711 250.261934,246.581411 C252.176407,247.971926 253.393381,250.234545 255.829549,251.104446 L255.829549,250.582732 C254.611442,249.016479 254.263282,246.754952 253.046308,245.015144 C251.307592,243.275341 249.566702,241.709083 247.826899,239.96928 C242.781025,233.1847 236.518133,227.268984 229.732547,222.397748 C224.166065,218.56995 211.986355,213.35055 209.724764,206.913051 C209.724764,206.913051 209.55014,206.73951 209.376604,206.56597 C213.204371,206.217787 217.727416,204.82618 221.381691,203.781623 C227.297466,202.215374 232.690457,202.563548 238.779677,200.998382 C241.562919,200.30203 244.347292,199.432129 247.130533,198.562222 L247.130533,196.997075 C244.00022,193.864532 241.737588,189.689684 238.431517,186.731778 C229.558965,179.075113 219.815379,171.595269 209.724764,165.332394 C204.330685,161.852792 197.371472,159.590151 191.630412,156.633369 C189.543536,155.587729 186.062797,155.06712 184.845823,153.327299 C181.713302,149.499523 179.973499,144.45476 177.711982,139.930579 C172.668329,130.360605 167.794863,119.749306 163.44541,109.658665 C160.315096,102.872993 158.400651,96.0872892 154.572862,89.8245233 C136.653166,60.2479054 117.167095,42.3281102 87.242308,24.7554574 C80.8048232,21.1023052 73.1503779,19.5360541 64.9730628,17.6227124 C60.6246649,17.4480683 56.2740469,17.1009966 51.9245072,16.9263525 C49.1412661,15.7082579 46.3569195,12.4033092 43.9207465,10.8370441 C34.0058755,4.5730961 8.42942301,-8.99607108 1.1220548,8.92370237 C-3.57563582,20.2324516 8.08126754,31.366477 12.0825475,37.1086812 C15.041536,41.1100178 18.8682195,45.6330512 20.9562053,50.1572463 C22.1742941,53.1129226 22.5213621,56.2465484 23.7394509,59.3779766 C26.523793,67.031339 29.1323922,75.5578744 32.7866404,82.691806 C34.7010855,86.3449577 36.7879657,90.1727422 39.2241297,93.477666 C40.6157457,95.3910055 43.0508086,96.2620126 43.5736286,99.3934363 C41.1385747,102.873029 40.9639284,108.092452 39.5711977,112.441992 C33.3083548,132.101492 35.7445143,156.4588 44.617071,170.89889 C47.4003121,175.247297 54.0124069,184.818421 62.8850316,181.164182 C70.7141415,178.032744 68.9743337,168.115626 71.2358604,159.41661 C71.7586894,157.327523 71.4105022,155.937017 72.4539446,154.545383 C72.4548508,154.718924 72.4539446,154.893561 72.4539446,154.893561 C74.8901041,159.764788 77.3251715,164.46251 79.5866891,169.333656 C84.980736,177.858025 94.3750434,186.731674 102.204103,192.647503 C106.381181,195.777808 109.686136,201.171877 114.905523,203.086313 L114.905523,202.563484 L114.557345,202.563484 C113.512801,200.997231 111.947645,200.301958 110.55602,199.083892 C107.424591,195.952463 103.943884,192.124665 101.50883,188.645081 C94.2025447,178.901486 87.7639408,168.115635 82.0228486,156.980496 C79.2396075,151.587555 76.8034481,145.671734 74.5419214,140.278811 C73.497378,138.189729 73.497378,135.059406 71.7575748,134.015973 C69.1478745,137.842647 65.3211911,141.148717 63.4067461,145.846417 C60.1028916,153.327344 59.7536034,162.548097 58.5355192,172.116947 C57.8402503,172.291598 58.187332,172.116947 57.8391493,172.465138 C52.2726627,171.072408 50.3582176,165.332394 48.2701955,160.460052 C43.0507905,148.107926 42.1808935,128.273684 46.7050387,114.008236 C47.923123,110.353988 53.142528,98.8717128 51.0545422,95.3921019 C50.0110998,92.0860273 46.5303924,90.1726969 44.6170529,87.5629967 C42.3555262,84.2569266 39.9193668,80.082065 38.35421,76.4278576 C34.1782383,66.6853811 32.0902615,55.898411 27.5672354,46.1548154 C25.4792496,41.6306665 21.8261069,36.9340837 18.8682195,32.7592063 C15.563264,28.0615284 11.9090067,24.7554574 9.29927023,19.1878506 C8.43046966,17.2745089 7.21128439,14.1430915 8.60290945,12.0550966 C8.95109211,10.6634847 9.64634733,10.1417599 11.0390689,9.79357497 C13.3005955,7.87912949 19.7380848,10.3152907 21.9995616,11.3587363 C28.4370509,13.9673251 33.8300058,16.4046087 39.2240527,20.0577513 C41.6591111,21.7975523 44.2688113,25.1036232 47.4002396,25.9735239 L51.0544833,25.9735239 C56.6220709,27.1905053 62.8849274,26.3216898 68.1043279,27.8868652 C77.3261638,30.843648 85.6758644,35.1942543 93.1579696,39.8919512 C115.95003,54.332049 134.738553,74.8626147 147.440063,99.3934454 C149.528049,103.39367 150.396845,107.04901 152.31129,111.22389 C155.965538,119.749365 160.488578,128.448376 164.14173,136.799218 C167.794872,144.975401 171.274474,153.327362 176.493861,160.113048 C179.104667,163.765094 189.542403,165.67953 194.240026,167.593975 C197.719632,169.159141 203.113711,170.551871 206.245112,172.465206 C212.159801,176.117243 218.075576,180.29433 223.643145,184.295646 C226.427474,186.382535 235.125402,190.733144 235.648231,194.21165 L235.648276,194.211632 L235.648276,194.211632 Z" fill="#00546B"></path>
            <path d="M58.1864892,43.0222644 C55.2286063,43.0222644 53.1417305,43.3715526 51.0537447,43.8932806 C51.0537447,43.8923744 51.0537447,44.0679225 51.0537447,44.2414633 L51.4019319,44.2414633 C52.794658,47.0247034 55.2286154,48.9391485 56.968414,51.3741978 C58.3611446,54.1574389 59.5781143,56.9417945 60.9708449,59.7261412 C61.1443766,59.5514948 61.3179175,59.3779585 61.3179175,59.3779585 C63.7551915,57.6370498 64.9721657,54.8538087 64.9721657,50.6789426 C63.9276177,49.4608583 63.7540769,48.2427786 62.8841798,47.0246944 C61.8407374,45.283782 59.5781052,44.414995 58.1864892,43.0222644 L58.1864892,43.0222644 L58.1864892,43.0222644 Z" fill="#00546B"></path>
        </svg>
        ),
        iconText: "MySQL",
        image1Dark:
            "/img/landing_slider/mysql_dark_large.png",
        image1Light:
            "/img/landing_slider/mysql_white_large.png",
        image2Dark:
            "/img/landing_slider/mysql_dark_small.png",
        image2Light:
            "/img/landing_slider/mysql_white_small.png",
        backgroundImage:
            "dark:bg-landing-sweet-spot-glow-yellow-dark bg-landing-sweet-spot-glow-yellow-light",
    },
    {
        title: "Object caching & Search",
        description: `REDIS, Memcached and Elasticsearch can be installed with a single click for each user individually.`,
        icon: (props: { active: boolean }) => (
        <svg width="24" height="24" viewBox="0 -18 256 256" xmlns="http://www.w3.org/2000/svg" preserveAspectRatio="xMinYMin meet">
            <path
                d="M245.97 168.943c-13.662 7.121-84.434 36.22-99.501 44.075-15.067 7.856-23.437 7.78-35.34 2.09-11.902-5.69-87.216-36.112-100.783-42.597C3.566 169.271 0 166.535 0 163.951v-25.876s98.05-21.345 113.879-27.024c15.828-5.679 21.32-5.884 34.79-.95 13.472 4.936 94.018 19.468 107.331 24.344l-.006 25.51c.002 2.558-3.07 5.364-10.024 8.988"
                fill="#912626"
            />
            <path
                d="M245.965 143.22c-13.661 7.118-84.431 36.218-99.498 44.072-15.066 7.857-23.436 7.78-35.338 2.09-11.903-5.686-87.214-36.113-100.78-42.594-13.566-6.485-13.85-10.948-.524-16.166 13.326-5.22 88.224-34.605 104.055-40.284 15.828-5.677 21.319-5.884 34.789-.948 13.471 4.934 83.819 32.935 97.13 37.81 13.316 4.881 13.827 8.9.166 16.02"
                fill="#C6302B"
            />
            <path
                d="M245.97 127.074c-13.662 7.122-84.434 36.22-99.501 44.078-15.067 7.853-23.437 7.777-35.34 2.087-11.903-5.687-87.216-36.112-100.783-42.597C3.566 127.402 0 124.67 0 122.085V96.206s98.05-21.344 113.879-27.023c15.828-5.679 21.32-5.885 34.79-.95C162.142 73.168 242.688 87.697 256 92.574l-.006 25.513c.002 2.557-3.07 5.363-10.024 8.987"
                fill="#912626"
            />
            <path
                d="M245.965 101.351c-13.661 7.12-84.431 36.218-99.498 44.075-15.066 7.854-23.436 7.777-35.338 2.087-11.903-5.686-87.214-36.112-100.78-42.594-13.566-6.483-13.85-10.947-.524-16.167C23.151 83.535 98.05 54.148 113.88 48.47c15.828-5.678 21.319-5.884 34.789-.949 13.471 4.934 83.819 32.933 97.13 37.81 13.316 4.88 13.827 8.9.166 16.02"
                fill="#C6302B"
            />
            <path
                d="M245.97 83.653c-13.662 7.12-84.434 36.22-99.501 44.078-15.067 7.854-23.437 7.777-35.34 2.087-11.903-5.687-87.216-36.113-100.783-42.595C3.566 83.98 0 81.247 0 78.665v-25.88s98.05-21.343 113.879-27.021c15.828-5.68 21.32-5.884 34.79-.95C162.142 29.749 242.688 44.278 256 49.155l-.006 25.512c.002 2.555-3.07 5.361-10.024 8.986"
                fill="#912626"
            />
            <path
                d="M245.965 57.93c-13.661 7.12-84.431 36.22-99.498 44.074-15.066 7.854-23.436 7.777-35.338 2.09C99.227 98.404 23.915 67.98 10.35 61.497-3.217 55.015-3.5 50.55 9.825 45.331 23.151 40.113 98.05 10.73 113.88 5.05c15.828-5.679 21.319-5.883 34.789-.948 13.471 4.935 83.819 32.934 97.13 37.811 13.316 4.876 13.827 8.897.166 16.017"
                fill="#C6302B"
            />
            <path
                d="M159.283 32.757l-22.01 2.285-4.927 11.856-7.958-13.23-25.415-2.284 18.964-6.839-5.69-10.498 17.755 6.944 16.738-5.48-4.524 10.855 17.067 6.391M131.032 90.275L89.955 73.238l58.86-9.035-17.783 26.072M74.082 39.347c17.375 0 31.46 5.46 31.46 12.194 0 6.736-14.085 12.195-31.46 12.195s-31.46-5.46-31.46-12.195c0-6.734 14.085-12.194 31.46-12.194"
                fill="#FFF"
            />
            <path d="M185.295 35.998l34.836 13.766-34.806 13.753-.03-27.52" fill="#621B1C" />
            <path d="M146.755 51.243l38.54-15.245.03 27.519-3.779 1.478-34.791-13.752" fill="#9A2928" />
        </svg>
        ),
        iconText: "REDIS",
        image1Dark:
            "/img/ilustrations/redis_large.png",
        image1Light:
            "/img/ilustrations/redis_large.png",
        image2Dark:
            "/img/ilustrations/redis_small.png",
        image2Light:
            "/img/ilustrations/redis_small.png",
        backgroundImage:
            "dark:bg-landing-sweet-spot-glow-cyan-dark bg-landing-sweet-spot-glow-cyan-light",
    },
    {
        title: "PHP, Python and NodeJS",
        description: `Each user can install PHP versions, edit .ini files and set versions per domain name.`,
        icon: (props: { active: boolean }) => (
        <svg viewBox="0 0 256 134" xmlns="http://www.w3.org/2000/svg" preserveAspectRatio="xMinYMin meet" width="24" height="24">
            <g fill-rule="evenodd">
                <ellipse fill="#8993BE" cx="128" cy="66.63" rx="128" ry="66.63" />
                <path
                    d="M35.945 106.082l14.028-71.014H82.41c14.027.877 21.041 7.89 21.041 20.165 0 21.041-16.657 33.315-31.562 32.438H56.11l-3.507 18.411H35.945zm23.671-31.561L64 48.219h11.397c6.137 0 10.52 2.63 10.52 7.89-.876 14.905-7.89 17.535-15.78 18.412h-10.52zM100.192 87.671l14.027-71.013h16.658l-3.507 18.41h15.78c14.028.877 19.288 7.89 17.535 16.658l-6.137 35.945h-17.534l6.137-32.438c.876-4.384.876-7.014-5.26-7.014H124.74l-7.89 39.452h-16.658zM153.425 106.082l14.027-71.014h32.438c14.028.877 21.042 7.89 21.042 20.165 0 21.041-16.658 33.315-31.562 32.438h-15.781l-3.507 18.411h-16.657zm23.67-31.561l4.384-26.302h11.398c6.137 0 10.52 2.63 10.52 7.89-.876 14.905-7.89 17.535-15.78 18.412h-10.521z"
                    fill="#232531"
                />
            </g>
        </svg>
        ),
        iconText: "PHP",
        image1Dark:
            "/img/ilustrations/php_large.png",
        image1Light:
            "/img/ilustrations/php_large.png",
        image2Dark:
            "/img/ilustrations/php_small.png",
        image2Light:
            "/img/ilustrations/php_small.png",
        backgroundImage:
            "dark:bg-landing-sweet-spot-glow-blue-dark bg-landing-sweet-spot-glow-blue-light",
    },
    {
        title: "ModSecurity & Firewall",
        description: `Each user can enable/disable ModSecurity for their websites or block IP addresses per domain name.`,
        icon: (props: { active: boolean }) => (
             <AuthenticationIcon
                className={clsx(
                    props.active
                        ? "dark:text-[#5959FF] text-[#693BC6]"
                        : "text-gray-500",
                )}
            />
        ),
        iconText: "Firewall",
        image1Dark:
            "/img/ilustrations/firewall_large.png",
        image1Light:
            "/img/ilustrations/firewall_large.png",
        image2Dark:
            "/img/ilustrations/firewall_small.png",
        image2Light:
            "/img/ilustrations/firewall_small.png",
        backgroundImage:
            "dark:bg-landing-sweet-spot-glow-indigo-dark bg-landing-sweet-spot-glow-indigo-light",
    },
];
