import clsx from "clsx";
import React, { useMemo } from "react";

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
		The features you need to build exactly what you want.
                </div>
                <div
                    className={clsx(
                        "font-semibold",
                        "text-gray-700 dark:text-gray-300",
                        "text-xs sm:text-base",
                        "mt-4 sm:mt-8",
                    )}
                >
		Tailored for today's hosts, OpenPanel provides a comprehensive set of tools and features you need to build and scale a cluster to meet your needs.
                </div>
            </div>

            <div
                className={clsx(
                    "my-10",
                    "border-b border-gray-200 dark:border-gray-700",
                )}
            />

            <Title>Web servers</Title>
            <List data={uiPackages} />

            <Title className="mt-20">User services</Title>
            <List data={dataProviderPackages} />

            <Title className="mt-20">
                Server management
            </Title>
            <List data={communityDataProviderPackages} />

            <Title className="mt-20">User interface</Title>
            <List data={frameworks} />

            <Title className="mt-20">User management</Title>
            <List data={integrations} />

            <Title className="mt-20">Security</Title>
            <List data={liveProviders} />

            <Title className="mt-20">Integrations</Title>
            <List data={communityPackages} />
        </IntegrationsLayout>
    );
};

export default Integrations;
