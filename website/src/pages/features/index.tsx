import clsx from "clsx";
import React, { useMemo } from "react";
import Translate from "@docusaurus/Translate";

import IntegrationsLayout from "@site/src/components/integrations/layout";
import { Integration } from "@site/src/types/integrations";
import { integrations as integrationsData } from "../../assets/integrations";
import Card from "../../components/integrations/card";

const Title = ({
    children,
    className,
}: React.PropsWithChildren<{ className?: string }>) => {
    return (
        <div
            className={clsx(
                "font-semibold",
                "text-gray-700 dark:text-gray-200",
                "text-base sm:text-2xl",
                className,
            )}
        >
            {children}
        </div>
    );
};

const List = ({ data }: { data: Integration[] }) => {
    return (
        <div
            className={clsx(
                "grid",
                "grid-cols-1 lg:grid-cols-2",
                "gap-8",
                "mt-8",
            )}
        >
            {data.map((item) => (
                <Card
                    key={item.name}
                    title={item.name}
                    description={item.description}
                    linkUrl={item.url}
                    icon={item.icon}
                    contributors={item.contributors}
                />
            ))}
        </div>
    );
};

const Integrations: React.FC = () => {
    const {
        communityPackages,
        dataProviderPackages,
        communityDataProviderPackages,
        frameworks,
        integrations,
        liveProviders,
        uiPackages,
    } = useMemo(() => {
        return {
            uiPackages: integrationsData["ui-framework-packages"],
            dataProviderPackages: integrationsData["data-provider-packages"],
            communityDataProviderPackages:
                integrationsData["community-data-provider-packages"],
            frameworks: integrationsData["frameworks"],
            integrations: integrationsData["integrations"],
            liveProviders: integrationsData["live-providers"],
            communityPackages: integrationsData["community-packages"],
        };
    }, []);

    return (
        <IntegrationsLayout>
            <div className={clsx("max-w-[624px]")}>
                <div
                    className={clsx(
                        "font-semibold",
                        "text-gray-700 dark:text-gray-200",
                        "text-xl sm:text-[40px] sm:leading-[56px]",
                    )}
                >
                    <Translate id="features.hero.title">Not quite what you would expect from a Control Panel</Translate>
                </div>
                <div
                    className={clsx(
                        "font-semibold",
                        "text-gray-700 dark:text-gray-300",
                        "text-xs sm:text-base",
                        "mt-4 sm:mt-8",
                    )}
                >
                    <Translate id="features.hero.subtitle.1">OpenPanel is arguebly the most customizable web hosting control panel.</Translate>
                </div>
                <div
                    className={clsx(
                        "font-semibold",
                        "text-gray-700 dark:text-gray-300",
                        "text-xs sm:text-base",
                        "mt-2 sm:mt-4",
                    )}
                >
                    <Translate id="features.hero.subtitle.2">Assign different mysql versions, web servers, and limits per user, let them run Docker containers, manage backups, schedule cron jobs in seconds, tweak configurations, and so much more.</Translate>
                </div>
            </div>

            <div
                className={clsx(
                    "my-10",
                    "border-b border-gray-200 dark:border-gray-700",
                )}
            />

            <Title><Translate id="features.section.1">Web server per user</Translate></Title>
            <List data={uiPackages} />

            <Title className="mt-20"><Translate id="features.section.2">All the classic features</Translate></Title>
            <List data={dataProviderPackages} />

            <Title className="mt-20"><Translate id="features.section.3">Smarter srver management</Translate></Title>
            <List data={communityDataProviderPackages} />

            <Title className="mt-20"><Translate id="features.section.4">No BS user interface</Translate></Title>
            <List data={frameworks} />

            <Title className="mt-20"><Translate id="features.section.5">Manage users easily</Translate></Title>
            <List data={integrations} />

            <Title className="mt-20"><Translate id="features.section.6">Built-in isolation and security</Translate></Title>
            <List data={liveProviders} />

            <Title className="mt-20"><Translate id="features.section.7">Integrate your billing panel</Translate></Title>
            <List data={communityPackages} />
        </IntegrationsLayout>
    );
};

export default Integrations;
