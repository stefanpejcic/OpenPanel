import React, { useState, useEffect } from "react";
import Head from "@docusaurus/Head";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";

const Assets: React.FC = () => {
    const [assets, setAssets] = useState([
        { name: "Logo Icon", svgUrl: "/img/svg/openpanel_logo.svg", svgCode: '', color: '#000000' },
        //{ name: "White Logo Icon", svgUrl: "/img/svg/openpanel_white_logo.svg", svgCode: '', color: '#FFFFFF' },
    ]);

    // Function to fetch and set initial SVG code for each asset
    useEffect(() => {
        assets.forEach((asset, index) => {
            fetch(asset.svgUrl)
                .then(response => response.text())
                .then(svgCode => {
                    const updatedAssets = [...assets];
                    updatedAssets[index].svgCode = svgCode;
                    setAssets(updatedAssets);
                });
        });
    }, []); // Empty dependency array ensures this only runs once on mount

    // Function to handle color change and debounce updates
    const handleColorChange = (color, index) => {
        const updatedAssets = [...assets];
        updatedAssets[index].color = color;
        setAssets(updatedAssets);
    };

    // Function to download SVG with selected color
    const downloadSVG = (svgCode, color, name) => {
        const coloredSvgCode = svgCode.replace(/fill="currentColor"/g, `fill="${color}"`);
        const blob = new Blob([coloredSvgCode], { type: 'image/svg+xml' });
        const url = URL.createObjectURL(blob);
        const link = document.createElement('a');
        link.href = url;
        link.download = `${name}.svg`;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
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
                                <div className="bg-[#f8f9fa] p-4 rounded-lg" dangerouslySetInnerHTML={{ __html: asset.svgCode.replace(/fill="currentColor"/g, `fill="${asset.color}"`) }} />
                                <div>
                                    Icon color: <input 
                                        type="color" 
                                        value={asset.color} 
                                        onChange={(e) => handleColorChange(e.target.value, index)} 
                                        className="mb-2"
                                    />
                                    <button onClick={() => downloadSVG(asset.svgCode, asset.color, asset.name)} className="bg-gray-500 text-white px-4 py-2">Download</button>
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
