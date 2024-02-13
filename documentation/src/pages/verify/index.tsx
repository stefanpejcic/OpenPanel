import React, { useState, useEffect } from "react";
import Head from "@docusaurus/Head";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";
import clsx from "clsx";

const Verify: React.FC = () => {
    const [ipAddress, setIpAddress] = useState("");
    const [responseData, setResponseData] = useState<any>(null);

    const handleCheckLicense = async () => {
        try {
            const response = await fetch(`https://verify.openpanel.co/?ip=${ipAddress}`);
            const data = await response.json();
            setResponseData(data);
        } catch (error) {
            console.error("Error fetching license data:", error);
        }
    };

    useEffect(() => {
        // Check if there's an 'ip' parameter in the URL
        const urlParams = new URLSearchParams(window.location.search);
        const urlIpAddress = urlParams.get('ip');

        if (urlIpAddress) {
            setIpAddress(urlIpAddress);
            handleCheckLicense(); // Automatically check the license if IP is present in the URL
        }
    }, []); // Empty dependency array ensures it only runs once on component mount

    return (
        <CommonLayout>
            <Head title="LICENSE | OpenPanel">
                <html data-page="support" data-customized="true" />
            </Head>
            <div className="refine-prose">
                <CommonHeader hasSticky={true} />

                <div className="flex-1 flex flex-col pt-8 lg:pt-16 pb-32 max-w-[800px] w-full mx-auto px-2">
                    <h1>License Verification</h1>

                    <div className="mb-4">
                        <label htmlFor="ipAddress">Enter IP Address:</label>
                        <input
                            type="text"
                            id="ipAddress"
                            value={ipAddress}
                            onChange={(e) => setIpAddress(e.target.value)}
                        />
                    </div>

                    <button className="bg-blue-500 text-white px-4 py-2" onClick={handleCheckLicense}>
                        Check License
                    </button>

                    {responseData && (
                        <div className="mt-4">
                            <h2>License Information:</h2>
                            <p>{responseData.message}</p>
                            {responseData.active_date && <p>Active Date: {responseData.active_date}</p>}
                        </div>
                    )}
                </div>
                <BlogFooter />
            </div>
        </CommonLayout>
    );
};

export default Verify;
