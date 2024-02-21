import React, { useEffect, useState } from "react";
import Head from "@docusaurus/Head";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";
import clsx from "clsx";

const Roadmap: React.FC = () => {
    const [milestones, setMilestones] = useState<any[]>([]);

    useEffect(() => {
        const fetchMilestones = async () => {
            try {
                const response = await fetch("https://api.github.com/repos/stefanpejcic/openpanel/milestones");
                const data = await response.json();
                setMilestones(data);
            } catch (error) {
                console.error("Error fetching milestones:", error);
            }
        };

        fetchMilestones();
    }, []);

    const calculateProgress = (milestone) => {
        const totalIssues = milestone.open_issues + milestone.closed_issues;
        const progress = (milestone.closed_issues / totalIssues) * 100;
        return progress.toFixed(2);
    };

    const formatDueDate = (dueDate) => {
        const options = { year: 'numeric', month: 'long', day: 'numeric' };
        return new Date(dueDate).toLocaleDateString(undefined, options);
    };

    return (
        <CommonLayout>
            <Head title="ROADMAP | OpenPanel">
                <html data-page="roadmap" data-customized="true" />
            </Head>
            <div className="refine-prose">
                <CommonHeader hasSticky={true} />

                <div className="flex-1 flex flex-col pt-8 lg:pt-16 pb-32 max-w-[800px] w-full mx-auto px-2">
                    <h1>Roadmap</h1>

                    <ul>
                        {milestones.map((milestone) => (
                            <li key={milestone.id}>
                                <strong>{milestone.title}</strong>
                                <p>{milestone.description}</p>
                                <p>Scheduled release date: {formatDueDate(milestone.due_on)}</p>
                                <p>Progress: {calculateProgress(milestone)}%</p>
                            </li>
                        ))}
                    </ul>
                </div>
                <BlogFooter />
            </div>
        </CommonLayout>
    );
};

export default Roadmap;
