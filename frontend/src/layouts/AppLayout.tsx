import { Outlet, NavLink, useNavigate } from "react-router-dom";
import { useEffect, useState } from "react";
import {
  Header,
  HeaderName,
  HeaderGlobalAction,
  HeaderGlobalBar,
  SideNav,
  SideNavItems,
  SideNavLink,
  Content,
} from "@carbon/react";
import { Logout, Notification } from "@carbon/icons-react";
import { useAuth } from "../auth/AuthContext";
import { unreadCount } from "../api/notifications";

export function AppLayout() {
  const { user, can, logout } = useAuth();
  const navigate = useNavigate();
  const [unread, setUnread] = useState(0);
  useEffect(() => {
    let active = true;
    async function poll() {
      if (document.visibilityState !== "visible") return;
      try {
        const count = await unreadCount();
        if (active) setUnread(count);
      } catch {
        // The notifications page surfaces persistent API errors.
      }
    }
    void poll();
    const timer = window.setInterval(poll, 45_000);
    return () => { active = false; window.clearInterval(timer) };
  }, []);
  return (
    <>
      <Header aria-label="Keycloak Onboarder">
        <HeaderName href="/" prefix="SSO">
          App Onboarder
        </HeaderName>

        <HeaderGlobalBar>
          <span className="header-user">{user.display_name || user.username} · {user.effective_role}</span>
          <HeaderGlobalAction aria-label={`${unread} unread notifications`} onClick={() => navigate("/notifications")}>
            <Notification size={20} />
            {unread > 0 && <span className="notification-count">{unread}</span>}
          </HeaderGlobalAction>
          <HeaderGlobalAction aria-label="Logout" onClick={logout}>
            <Logout size={20} />
          </HeaderGlobalAction>
        </HeaderGlobalBar>
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
          <SideNavLink as={NavLink} to="/approvals">Approvals</SideNavLink>
          <SideNavLink as={NavLink} to="/notifications">Notifications</SideNavLink>
          {can("view_audit") && <SideNavLink as={NavLink} to="/audit-logs">Audit Log</SideNavLink>}
          {can("manage_settings") && <SideNavLink as={NavLink} to="/settings">Settings</SideNavLink>}
        </SideNavItems>
      </SideNav>

      <Content className="app-content">
        <Outlet />
      </Content>
    </>
  );
}
