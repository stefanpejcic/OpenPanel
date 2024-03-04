import React, { useState, useEffect } from "react";
import Head from "@docusaurus/Head";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";

const Assets: React.FC = () => {
    const [color, setColor] = useState("#000000"); // Default color
    const [assets, setAssets] = useState([
        { name: "Logo Icon", svgUrl: "/img/svg/openpanel_logo.svg", svgContent: null },
        //{ name: "White Logo Icon", svgUrl: "/img/svg/openpanel_white_logo.svg", svgContent: null },
    ]);

    // Fetch SVG content for each asset
    useEffect(() => {
        assets.forEach((asset, index) => {
            fetch(asset.svgUrl)
                .then(response => response.text())
                .then(svgContent => {
                    const updatedAssets = [...assets];
                    updatedAssets[index] = { ...asset, svgContent: svgContent.replace(/fill="currentColor"/g, `fill="${color}"`) };
                    setAssets(updatedAssets);
                });
        });
    }, [color]); // Re-fetch when color changes

    const handleCopySVGCode = async (svgContent: string) => {
        const modifiedSVG = svgContent.replace(/fill="#[0-9A-Fa-f]{6}"/g, `fill="${color}"`);
        navigator.clipboard.writeText(modifiedSVG).then(
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
                                <div className="bg-[#f8f9fa] p-4 rounded-lg" dangerouslySetInnerHTML={{ __html: asset.svgContent }} />
                                <div>
                                    <a href={asset.svgUrl} target="_blank" rel="noopener noreferrer" className="block bg-blue-500 text-white px-4 py-2 mb-2">Download</a>
                                    <input 
                                        type="color" 
                                        value={color} 
                                        onChange={(e) => setColor(e.target.value)} 
                                        className="mb-2"
                                    />
                                    <button onClick={() => handleCopySVGCode(asset.svgContent)} className="bg-gray-500 text-white px-4 py-2">Copy SVG</button>
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
