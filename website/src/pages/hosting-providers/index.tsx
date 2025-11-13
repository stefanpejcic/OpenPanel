import React from "react";
import { CommonLayout } from "@site/src/refine-theme/common-layout";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";
import Head from "@docusaurus/Head";
import clsx from "clsx";

const providers = [
  {
    company: "HOSTKEY",
    location: "EU",
    services: "VPS, Dedicated",
    managedSupport: "No",
    freeEnterprise: "No",
    link: "https://hostkey.com/apps/hosting-control-panels/openpanel/",
    logo: "https://cdn.hostadvice.com/2016/04/hostkey_logo-250x75.png",
  },
  {
    company: "UNLIMITED.RS",
    location: "EU",
    services: "VPS",
    managedSupport: "No",
    freeEnterprise: "No",
    link: "https://unlimited.rs/vps-hosting/",
    logo: "https://unlimited.rs/wp-content/themes/unlimited-RS/themeFunctions/media/unlimited.rs_logo.svg",
  },
  {
    company: "Hostigan",
    location: "EU",
    services: "VPS, Dedicated",
    managedSupport: "Yes",
    freeEnterprise: "Yes",
    link: "https://hostigan.com/openpanel/",
    logo: "https://hostigan.com/wp-content/uploads/2025/04/s-768x135.png",
  },
  {
    company: "AltusHost",
    location: "EU",
    services: "VPS, Shared",
    managedSupport: "No",
    freeEnterprise: "No",
    link: "https://altushost.com",
    logo: "https://www.altushost.com/wp-content/themes/altushost/themeFunctions/media/altushost.svg",
  }, 
  {
    company: "Clouding.io",
    location: "EU",
    services: "VPS",
    managedSupport: "No",
    freeEnterprise: "No",
    link: "https://clouding.io",
    logo: "/img/svg/clouding.io-svg-logo.png",
  },  
];

const HostingProvidersPage: React.FC = () => {
  return (
    <>
      <Head title="Control Panel for Hosting Providers | OpenPanel Enterprise">
        <html data-page="partners-table" />
      </Head>

      <div className="refine-prose">
        <CommonHeader hasSticky={true} />

        <div
          className={clsx(
            "not-prose",
            "xl:max-w-[944px] xl:py-16",
            "lg:max-w-[912px] lg:py-10",
            "md:max-w-[624px] md:text-4xl md:pb-6 pt-6",
            "sm:max-w-[480px] text-xl",
            "max-w-[328px]",
            "w-full mx-auto",
          )}
        >
          <h1
            className={clsx(
              "font-semibold",
              "!mb-0",
              "text-gray-900 dark:text-gray-0",
              "text-xl md:text-[40px] md:leading-[56px]",
            )}
          >
            OpenPanel{" "}
            <span className="text-transparent bg-clip-text bg-gradient-to-r from-[#0FBDBD] to-[#26D97F]">
              Hosting Partners
            </span>
          </h1>
        </div>

        <div
          className={clsx(
            "max-w-[912px] w-full mx-auto py-4",
            "text-center text-gray-800 dark:text-gray-200"
          )}
        >
          <p className="text-left">
            If you want to be on the OpenPanel Hosting Partners list:{" "}
            <a
              href="/licenses-for-partners"
              rel="noopener noreferrer"
              className="text-blue-600 dark:text-blue-400 underline"
            >
              OpenPanel NOC Partner package
            </a>
          </p>

          <div className="mt-8 grid grid-cols-1 md:grid-cols-2 gap-6">
            {providers.map((p, idx) => (
              <div
                key={idx}
                className="border border-gray-200 dark:border-gray-700 rounded-2xl p-4 shadow-sm bg-white dark:bg-gray-900 flex flex-col justify-between"
              >
                <div className="text-center">
                  {p.logo ? (
                    <img
                      src={p.logo}
                      alt={p.company}
                      className="h-10 mb-3 object-contain mx-auto"
                    />
                  ) : (
                    <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-0 mx-auto">
                      {p.company}
                    </h3>
                  )}

                  <ul className="mt-2 space-y-1 text-sm text-gray-700 dark:text-gray-300 text-left">
                    <li>
                      <strong>Server Location:</strong> {p.location}
                    </li>
                    <li>
                      <strong>Services:</strong> {p.services}
                    </li>
                    <li>
                      <strong>Managed Support:</strong> {p.managedSupport}
                    </li>
                    <li>
                      <strong>Free Enterprise license:</strong> {p.freeEnterprise}
                    </li>
                  </ul>
                </div>
                <div className="mt-4">
                  <a
                    href={p.link}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="inline-block px-4 py-2 text-sm font-medium text-gray-900 dark:text-gray-0 bg-gradient-to-r from-[#0FBDBD] to-[#26D97F] rounded-xl hover:opacity-90 transition"
                  >
                    Visit {p.company}
                  </a>
                </div>
              </div>
            ))}
          </div>
        </div>

        <BlogFooter />
      </div>
    </>
  );
};

export default function ProvidersTable() {
  return (
    <CommonLayout>
      <HostingProvidersPage />
    </CommonLayout>
  );
}
