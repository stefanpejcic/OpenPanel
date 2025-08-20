import React, { useState, useEffect } from "react";
import Head from "@docusaurus/Head";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";

interface ApiData {
  active_install: number;
  recent_updates: number;
  latest_version: string;
  last_updated: string;
}

const Stats: React.FC = () => {
  const [data, setData] = useState<ApiData | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const response = await fetch(
          "https://usage-api.openpanel.org/active_install"
        );
        const json = await response.json();
        setData(json);
      } catch (err) {
        console.error("Failed to fetch data:", err);
      }
    };
    fetchData();
  }, []);

  // Format the last_updated date
  const formattedDate = data
    ? new Date(data.last_updated).toLocaleDateString(undefined, {
        year: "numeric",
        month: "short",
        day: "numeric",
      })
    : "";

  return (
    <CommonLayout>
      <Head title="Usage Statistics | OpenPanel">
        <html data-page="docs" data-customized="true" />
      </Head>

      <div className="refine-prose">
        <CommonHeader hasSticky={true} />

        <div className="not-prose xl:max-w-[944px] xl:py-16 lg:max-w-[912px] lg:py-10 md:max-w-[624px] md:text-4xl md:pb-6 pt-6 sm:max-w-[480px] text-xl max-w-[328px] w-full mx-auto">
          <div
            id="playground"
            className="flex flex-col gap-8 landing-sm:gap-12 landing-md:gap-8"
            style={{ scrollMarginTop: "6rem" }}
          >
            <div className="w-full rounded-2xl landing-md:rounded-3xl relative overflow-hidden transition-[min-height,height] duration-300 ease-out min-h-[515px]">
              {data && (
                <div className="grid grid-cols-1 landing-sm:grid-cols-2 gap-4 landing-sm:gap-6">
                  <a 
                    href="https://openpanel.com/docs/changelog/{data.latest_version}"
                    className="block not-prose p-4 landing-sm:py-4 landing-sm:px-10 dark:bg-landing-noise dark:bg-gray-800 bg-gray-50 rounded-2xl landing-sm:rounded-3xl no-underline">
                    <div className="whitespace-nowrap text-[40px] leading-[48px] landing-sm:text-[64px] landing-sm:leading-[72px] dark:bg-landing-stats-text-dark bg-landing-stats-text bg-clip-text text-transparent font-bold drop-shadow-2xl">
                      {data.latest_version}
                    </div>
                    <div className="mt-2 landing-sm:mt-6 text-base dark:text-gray-400 text-gray-600">
                      Latest Version
                    </div>
                  </a>

                  <a className="block not-prose p-4 landing-sm:py-4 landing-sm:px-10 dark:bg-landing-noise dark:bg-gray-800 bg-gray-50 rounded-2xl landing-sm:rounded-3xl no-underline">
                    <div className="whitespace-nowrap text-[40px] leading-[48px] landing-sm:text-[64px] landing-sm:leading-[72px] dark:bg-landing-stats-text-dark bg-landing-stats-text bg-clip-text text-transparent font-bold drop-shadow-2xl">
                      {formattedDate}
                    </div>
                    <div className="mt-2 landing-sm:mt-6 text-base dark:text-gray-400 text-gray-600">
                      Latest Update
                    </div>
                  </a>

                  <a
                    className="block not-prose p-4 landing-sm:py-4 landing-sm:px-10 dark:bg-landing-noise dark:bg-gray-800 bg-gray-50 rounded-2xl landing-sm:rounded-3xl no-underline"
                  >
                    <div className="whitespace-nowrap text-[40px] leading-[48px] landing-sm:text-[64px] landing-sm:leading-[72px] dark:bg-landing-stats-text-dark bg-landing-stats-text bg-clip-text text-transparent font-bold drop-shadow-2xl">
                      {data.active_install}
                    </div>
                    <div className="mt-2 landing-sm:mt-6 text-base dark:text-gray-400 text-gray-600">
                      Active Installations
                    </div>
                  </a>

                  <a 
                    href="/docs/changelog/intro/"
                    className="block not-prose p-4 landing-sm:py-4 landing-sm:px-10 dark:bg-landing-noise dark:bg-gray-800 bg-gray-50 rounded-2xl landing-sm:rounded-3xl no-underline">
                    <div className="whitespace-nowrap text-[40px] leading-[48px] landing-sm:text-[64px] landing-sm:leading-[72px] dark:bg-landing-stats-text-dark bg-landing-stats-text bg-clip-text text-transparent font-bold drop-shadow-2xl">
                      {data.recent_updates}
                    </div>
                    <div className="mt-2 landing-sm:mt-6 text-base dark:text-gray-400 text-gray-600">
                      Updates in last 30 days
                    </div>
                  </a>
                </div>
              )}
            </div>
          </div>
        </div>

        <BlogFooter />
      </div>
    </CommonLayout>
  );
};

export default Stats;
