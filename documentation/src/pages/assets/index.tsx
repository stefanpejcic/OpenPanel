import React from "react";
import Head from "@docusaurus/Head";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";
import clsx from "clsx";

const Assets: React.FC = () => {
    const assets = [
        { name: "Logo Icon", svgUrl: "/img/svg/openpanel_logo.svg", downloadUrl: "/img/svg/openpanel_logo.svg" },
        { name: "White Logo Icon", svgUrl: "/img/svg/openpanel_white_logo.svg", downloadUrl: "/img/svg/openpanel_white_logo.svg" },
        // more here..
    ];


    const handleCopySVGCode = async (svgUrl: string) => {
        const response = await fetch(svgUrl);
        const svgCode = await response.text();
        navigator.clipboard.writeText(svgCode).then(
            () => console.log('SVG code copied to clipboard'),
            (err) => console.error('Error copying SVG code: ', err)
        );
    };

    return (
        <CommonLayout>
            <Head title="BRAND ASSETS | OpenPanel">
                <html data-page="brand-assets" data-customized="true" />
            </Head>
            <div className="refine-prose">
                <CommonHeader hasSticky={true} />

                <div className="flex-1 flex flex-col pt-8 lg:pt-16 pb-32 max-w-[800px] w-full mx-auto px-2">
                    <h1>Brand Assets</h1>

                    <div className="grid grid-cols-2 gap-4">
                        {assets.map((asset, index) => (
                            <div key={index} className="flex items-center space-x-4">
                                <div className="bg-[#f8f9fa] p-4 rounded-lg">
                                    <img src={asset.svgUrl} alt={asset.name} className="w-20 h-20" />
                                </div>
                                <div>
                                    <a href={asset.downloadUrl} target="_blank" rel="noopener noreferrer" className="block bg-blue-500 text-white px-4 py-2 mb-2">Download</a>
                                    <button onClick={() => handleCopySVGCode(asset.svgUrl)} className="bg-gray-500 text-white px-4 py-2">Copy SVG</button>
                                </div>
                            </div>
                        ))}
                    </div>
                </div>
                <BlogFooter />
            </div>
        </CommonLayout>
    );
};

export default Assets;
