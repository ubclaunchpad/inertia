import React from 'react';
import PropTypes from 'prop-types';
import FooterLaunchpad from '../../assets/logo/launchpad-logo-light-blue-01.svg';

import './index.sass';

const backgroundColor = '#474d5e';

const Footer = ({ version }) => {
  const inertiadVersion = `inertiad ${version}`;
  const icons = [
    {
      id: 1,
      link: 'https://github.com/ubclaunchpad/inertia',
      name: 'fab fa-github',
    },
    {
      id: 2,
      link: 'https://github.com/ubclaunchpad/inertia/issues/new/choose',
      name: 'fas fa-comments',
    },
    {
      id: 3,
      link: 'https://medium.com/ubc-launch-pad-software-engineering-blog',
      name: 'fab fa-medium-m',
    },
  ].map(socialFooterIcon => (
    <span key={socialFooterIcon.id} className="icons">
      <a
        href={socialFooterIcon.link}
        rel="noopener noreferrer"
        target="_blank">
        <span className={`${socialFooterIcon.name} pad-right-xs hover-color-lightred`} />
      </a>
    </span>
  ));
  return (
    <footer className="footer flex flow-row ai-center jc-between fill-width" style={{ backgroundColor }}>
      <span className="pad-left-xs">
        <p className="daemon-version">
          {inertiadVersion}
        </p>
      </span>
      <span className="pad-right-xs">
        {icons}
        <span className="icons">
          <div className="fill-height vertical-middle">
            <a href="https://www.ubclaunchpad.com" rel="noopener noreferrer" target="_blank">
              <FooterLaunchpad className=" hover-fill-lightred launchpad" />
            </a>
          </div>
        </span>
      </span>
    </footer>
  );
};

Footer.propTypes = {
  version: PropTypes.string,
};

export default Footer;
