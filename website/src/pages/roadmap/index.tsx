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
                    <p>Below is a glimpse of the features planned for OpenPanel.</p>
                    <p>We share this roadmap to give our users a clear view of our future direction and to encourage feedback on which features resonate the most (or least).</p>
                    <p>You can <a href="https://roadmap.openpanel.com/" target="_blank">suggest new features</a>.</p>
                    <ul>
                        <li>Further enhancements to cluster mode and OpenAdmin.</li>
                        <li>A complete email management UI.</li>
                        <li>Reseller account functionality.</li>
                        <li>Support for multiple Node.js and Python versions.</li>
                        <li>Autoinstallers for 40+ CMSs/scripts.</li>
                        <li>Backup and restore support, including S3 integration and partial backups.</li>
                        <li>Extensions for Blesta and Paymenter.org integration.</li>
                        <li>Support for additional Linux distributions.</li>
                        <li>Ongoing improvements in container isolation.</li>
                    </ul>
                    <h3>Planned releases</h3>
                    <ul>
                        {milestones.map((milestone) => (
                            <li key={milestone.id}>
                                <strong><a href={milestone.html_url} target="_blank" rel="noopener noreferrer">{milestone.title}</a></strong>
                                <p>{milestone.description}</p>
                                <p>Scheduled release: <strong>{formatDueDate(milestone.due_on)}</strong></p>
                                <p>Progress: {calculateProgress(milestone)}%</p>
                                
                                {/* Progress bar */}
                                <div style={{ width: '100%', height: '8px', backgroundColor: '#ddd', borderRadius: '4px', marginTop: '8px', overflow: 'hidden' }}>
                                    <div style={{ height: '100%', backgroundColor: '#4caf50', width: `${calculateProgress(milestone)}%`, transition: 'width 0.3s ease-in-out' }}></div>
                                </div>
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
