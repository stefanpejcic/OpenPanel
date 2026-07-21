import React from "react";
import Head from "@docusaurus/Head";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";

import BenchmarkContent from "@site/src/pages/calculator/benchmark/benchmark.md";

const CalculatorBenchmark: React.FC = () => {
    return (
        <CommonLayout description="How we benchmarked OpenPanel 2.0 (Podman) resource usage: test setup, the account-creation script used, and the raw opencli docker-collect_stats results.">
            <Head title="Benchmark Methodology | OpenPanel">
                <html data-page="calculator-benchmark" data-customized="true" />
            </Head>
            <div className="refine-prose">
                <CommonHeader hasSticky={true} />

                <div className="flex-1 flex flex-col pt-8 lg:pt-16 pb-32 max-w-[800px] w-full mx-auto px-2">
                    <BenchmarkContent />
                </div>
                <BlogFooter />
            </div>
        </CommonLayout>
    );
};

export default CalculatorBenchmark;
