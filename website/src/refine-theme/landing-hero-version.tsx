import React, { useState, useEffect } from 'react';
import { OpenPanelLogoIcon } from './icons/small-openpanel-logo';
import { HeroBadge } from './hero-badge';

export const LandingHeroVersion = () => {
    const [version, setVersion] = useState('2.0.4');

    useEffect(() => {
        const fetchVersion = async () => {
            try {
                const response = await fetch('https://api.openpanel.com/v2/statistics/latest_version');
                const data = await response.json();

                if (data && data.latest_version) {
                    setVersion(data.latest_version);
                }
            } catch (error) {
                console.error('Failed to fetch version:', error);
            }
        };

        fetchVersion();
    }, []);

    return (
        <HeroBadge
            href={`/docs/changelog/${version}`}
            icon={<OpenPanelLogoIcon className="text-refine-orange drop-shadow-none dark:drop-shadow-github-stars-glow" />}
        >
            <span className="font-semibold">
                latest version <span>{version}</span>
            </span>{" "}
            <span>- view the changelog</span>
        </HeroBadge>
    );
};
