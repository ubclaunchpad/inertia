import React from 'react';
import PropTypes from 'prop-types';
import './index.sass';

import InertiaLogo from '../../assets/logo/inertia-horizontal-black.png';
import DashboardIcon from '../../assets/icons/dashboard-icon.png';
import ContainerIcon from '../../assets/icons/container-icon.png';
import TeamIcon from '../../assets/icons/team-icon.png';
import SettingIcon from '../../assets/icons/settings-icon.png';

const NavBar = () => {
  const icons = [
    {
      id: 1,
      link: 'https://github.com/ubclaunchpad/inertia',
      name: 'fab fa-github',
      icon: DashboardIcon,
    },
    {
      id: 2,
      link: 'https://medium.com/ubc-launch-pad-software-engineering-blog',
      name: 'fab fa-medium-m',
      icon: ContainerIcon,
    },
    {
      id: 3,
      link: 'https://github.com/ubclaunchpad/inertia/issues/new/choose',
      name: 'fas fa-comments',
      icon: TeamIcon,
    },
    {
      id: 4,
      link: 'https://github.com/ubclaunchpad/inertia/issues/new/choose',
      name: 'fas fa-comments',
      icon: SettingIcon,
    },
  ].map(socialFooterIcon => (
    <span className="icons">
      <a key={socialFooterIcon.id} href={socialFooterIcon.link}>
        <span className={socialFooterIcon.name} />
      </a>
    </span>));
  return (
    <footer className="footer">
      <span className="left-side"><p className="daemon-version">Daemon</p></span>
      <span className="right-side">
        {icons}
        <a href="https://www.ubclaunchpad.com">
          <FooterLaunchpad />
          {/* <span className="launchpad"><img src={footerLaunchpad} height="17" width="105" alt="ubc launchpad" /></span> */}
        </a>
      </span>
    </footer>
  );
};

NavBar.propTypes = {
  version: PropTypes.string,
};

export default NavBar;
