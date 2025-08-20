import React, { useState, useEffect } from "react";
import Head from "@docusaurus/Head";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";

const Demo: React.FC = () => {
  const [version, setVersion] = useState("8000"); // default/fallback

  useEffect(() => {
    const fetchVersion = async () => {
      try {
        const response = await fetch(
          "https://usage-api.openpanel.org/active_install"
        );
        const data = await response.json();
        if (data && data.active_install) {
          setVersion(data.active_install);
        }
      } catch (err) {
        console.error("Failed to fetch data:", err);
      }
    };
    fetchVersion();
  }, []);

  return (
    <CommonLayout>
      <Head title="Active Installations Counter | OpenPanel">
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
              <div className="relative flex flex-col landing-md:flex-row w-full transition-[transform,opacity,margin-bottom] duration-300 ease-in-out">
                <div className="flex-1 rounded-2xl landing-md:rounded-3xl landing-md:rounded-tr-none landing-md:rounded-br-none flex flex-col gap-6 landing-sm:gap-10 pt-4 landing-sm:pt-10 landing-md:pt-16 px-4 landing-sm:px-10 pb-14 landing-sm:pb-20 landing-md:pb-16 bg-gray-50 dark:bg-gray-800 landing-md:bg-landing-wizard-option-bg-light dark:landing-md:bg-landing-wizard-option-bg-dark landing-md:bg-landing-wizard-option-left landing-md:bg-landing-wizard-option"
                     style={{ backgroundRepeat: "no-repeat, repeat" }}>
                  <p className="text-[32px] leading-[40px] tracking-[-0.5%] landing-sm:text-[56px] landing-sm:leading-[72px] landing-sm:tracking-[-2%] font-extrabold text-gray-900 dark:text-gray-0">
                    {version}
                  </p>
                  <div class="text-[24px] landing-sm:mt-6 text-base dark:text-gray-400 text-gray-600">
                      Active Installations 🎉
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <BlogFooter />
      </div>
    </CommonLayout>
  );
};

export default Demo;
