import clsx from "clsx";
import React, { FC, useEffect, useLayoutEffect, useRef } from "react";
import {
    HostkeyIcon,
    AltusHostIcon,
    DigitalOceanIcon,
} from "../components/landing/icons";
import { useInView } from "framer-motion";

type Props = {
    className?: string;
};

export const LandingTrustedByDevelopers: FC<Props> = ({ className }) => {
    const ref = useRef<HTMLDivElement>(null);
    const inView = useInView(ref);

    const lastChangedIndex = React.useRef<number>(0);

    const [randomIcons, setRandomIcons] = React.useState<IList>([]);

    useLayoutEffect(() => {
        setRandomIcons(list.sort(() => 0.5 - Math.random()).slice(0, 6));
    }, []);

    useEffect(() => {
        let interval: NodeJS.Timeout;
        // change one random icon in the list every X seconds.
        if (inView) {
            interval = setInterval(() => {
                setRandomIcons((prev) => {
                    const { changedIndex, newList } = changeOneRandomIcon(
                        prev,
                        list,
                        lastChangedIndex.current,
                    );
                    lastChangedIndex.current = changedIndex;
                    return newList;
                });
            }, 2000);
        }

        return () => clearInterval(interval);
    }, [randomIcons, inView]);

    return (
        <div className={clsx(className, "w-full")} ref={ref}>
            <div
                className={clsx(
                    "not-prose",
                    "relative",
                    "w-full",
                    "p-4 landing-md:p-10",
                    "dark:bg-landing-trusted-by-developers-dark bg-landing-trusted-by-developers",
                    "dark:bg-gray-800 bg-gray-50",
                    "rounded-2xl landing-sm:rounded-3xl",
                )}
            >
                <p
                    className={clsx(
                        "whitespace-nowrap",
                        "px-0 landing-sm:px-6 landing-lg:px-0",
                        "text-base landing-sm:text-2xl",
                        "dark:text-gray-400 text-gray-600",
                    )}
                >
                    Trusted by major hosting brands
                </p>
                <div
                    className={clsx(
                        "grid",
                        "grid-cols-3 landing-lg:grid-cols-6",
                        "min-h-[160px] landing-lg:min-h-[80px]",
                        "justify-center",
                        "items-center",
                        "mt-6",
                    )}
                >
                    {randomIcons.map((item) => (
                        <div
                            key={item.id}
                            className={clsx(
                                "max-w-[187px] w-full",
                                "overflow-hidden",
                            )}
                        >
                            <div
                                className={clsx(
                                    "animate-opacity-reveal",
                                    "flex",
                                    "items-center",
                                    "justify-center",
                                    "max-w-[187px]",
                                )}
                            >
                            <a href={item.href} target="_blank" rel="noopener noreferrer">
                                {item.icon}
                            </a>
                            </div>
                        </div>
                    ))}
                </div>
            </div>
        </div>
    );
};

// change only one random icon in the list
const changeOneRandomIcon = (
    currentList: IList,
    list: IList,
    lastChangedIndex: number,
): { newList: IList; changedIndex: number } => {
    const newList = [...currentList];

    // pick randomIndex from the current list
    let randomIndex = Math.floor(Math.random() * newList.length);
    // if the randomIndex is the same as the last changed index, pick another randomIndex
    while (randomIndex === lastChangedIndex) {
        randomIndex = Math.floor(Math.random() * newList.length);
    }

    // pick randomIcon from the list
    let randomIcon = list[Math.floor(Math.random() * list.length)];
    // check if the randomIcon is already in the current list
    let isExist = newList.find((item) => item.id === randomIcon.id);
    // if the randomIcon is already in the current list, pick another randomIcon
    while (isExist) {
        randomIcon = list[Math.floor(Math.random() * list.length)];
        isExist = newList.find((item) => item.id === randomIcon.id);
    }

    // change the randomIcon in the current list
    newList[randomIndex] = randomIcon;

    return { newList, changedIndex: randomIndex };
};

type IList = {
    icon: React.ReactNode;
    id: number;
    href: string;
}[];

const list: IList = [
    { icon: <HostkeyIcon />, id: 1, href: "https://hostkey.com/apps/hosting-control-panels/openpanel/" },
    { icon: <AltusHostIcon />, id: 2, href: "https://www.altushost.com/linux-vps/" },
    { icon: <DigitalOceanIcon />, id: 3, href: "https://digitalocean.com/" },
];
