import { Outlet, NavLink } from "react-router-dom";
import {
  Header,
  HeaderName,
  SideNav,
  SideNavItems,
  SideNavLink,
  Content,
} from "@carbon/react";

export function AppLayout() {
  return (
    <>
      <Header aria-label="Keycloak Onboarder">
        <HeaderName href="/" prefix="SSO">
          App Onboarder
        </HeaderName>
      </Header>

      <SideNav expanded isFixedNav aria-label="Side navigation">
        <SideNavItems>
          <SideNavLink as={NavLink} to="/dashboard">
            Dashboard
          </SideNavLink>
          <SideNavLink as={NavLink} to="/applications">
            Applications
          </SideNavLink>
          <SideNavLink as={NavLink} to="/templates">
            Templates
          </SideNavLink>
          <SideNavLink as={NavLink} to="/jobs">
            Provisioning Jobs
          </SideNavLink>
          <SideNavLink as={NavLink} to="/settings">
            Settings
          </SideNavLink>
        </SideNavItems>
      </SideNav>

      <Content className="app-content">
        <Outlet />
      </Content>
    </>
  );
}
