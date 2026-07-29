import React, { useState, useEffect } from 'react';
import { OrangeStarIcon } from './icons/orange-star';
import { HeroBadge } from './hero-badge';

const REPO = 'stefanpejcic/OpenPanel';

export const LandingHeroGithubStars = () => {
    const [stars, setStars] = useState<number | null>(null);

    useEffect(() => {
        const fetchStars = async () => {
            try {
                const response = await fetch(`https://api.github.com/repos/${REPO}`);
                const data = await response.json();

                if (data && typeof data.stargazers_count === 'number') {
                    setStars(data.stargazers_count);
                }
            } catch (error) {
                console.error('Failed to fetch GitHub stars:', error);
            }
        };

        fetchStars();
    }, []);

    return (
        <HeroBadge
            href={`https://github.com/${REPO}`}
            icon={<OrangeStarIcon className="drop-shadow-none dark:drop-shadow-github-stars-glow" />}
        >
            <span className="font-semibold">
                {stars !== null ? stars.toLocaleString() : "…"} stars
            </span>{" "}
            <span>on GitHub</span>
        </HeroBadge>
    );
};
