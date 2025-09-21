import React from "react";
import { CommonLayout } from "@site/src/refine-theme/common-layout";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";
import Head from "@docusaurus/Head";
import clsx from "clsx";

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
            OpenPanel Enterprise pricing for{" "}
            <span className="text-transparent bg-clip-text bg-gradient-to-r from-[#0FBDBD] to-[#26D97F]">
              NOC partners
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
            If you want to be on the OpenPanel NOC Partner Site list:{" "}
            <a href="https://my.openpanel.com/index.php?rp=/store/partners/noc" target="_blank">
              OpenPanel NOC Partner package
            </a>
          </p>
          <div className="overflow-x-auto mt-8">
            <table className="w-full text-left">
              <thead>
                <tr className="bg-gray-100 dark:bg-gray-800">
                  <th className="border px-4 py-2">Company</th>
                  <th className="border px-4 py-2">Server Location</th>
                  <th className="border px-4 py-2">Services</th>
                  <th className="border px-4 py-2">Managed Support</th>
                  <th className="border px-4 py-2">Free Enterprise license</th>
                </tr>
              </thead>
              <tbody>
                <tr><td>Hostkey</td><td>EU</td><td>VPS, Dedicated</td><td>No</td><td>No</td></tr>
                <tr><td>Hostigan</td><td>EU</td><td>VPS, Dedicated</td><td>Yes</td><td>Yes</td></tr>
              </tbody>
            </table>
          </div>
        </div>

        {/* Footer */}
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
