import { useState } from "react";
import { Outlet, useNavigate, useLocation } from "react-router-dom";
import {
  Box,
  Drawer,
  AppBar,
  Toolbar,
  List,
  Typography,
  ListItem,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  IconButton,
  Divider
} from "@mui/material";
import {
  PhoneAndroid,
  Webhook,
  Logout,
  Menu as MenuIcon,
} from "@mui/icons-material";
import KeyOutlined from "@ant-design/icons/KeyOutlined";
import User from "@ant-design/icons/UserOutlined";
import { clearScopes, getScopes } from "../utils/scopes";
import { message } from "antd";

const drawerWidth = 240;

const hasResourceScope = (resource) => {
  const userScopes = getScopes();
  return userScopes.some((s) => s === "*" || s.startsWith(resource + ":"));
};

const menuItems = [
  { text: "Tokens", path: "/tokens", icon: <KeyOutlined />, resource: "tokens" },
  { text: "Devices", path: "/devices", icon: <PhoneAndroid />, resource: "devices" },
  { text: "Webhooks", path: "/webhooks", icon: <Webhook />, resource: "webhooks" },
  { text: "Credentials", path: "/credentials", icon: <User />, resource: "credentials" },
];

export default function Layout() {
  const [mobileOpen, setMobileOpen] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();

  const handleDrawerToggle = () => {
    setMobileOpen(!mobileOpen);
  };

  const handleLogout = async () => {
    try {
      await fetch("/api/v1/admin/logout", { method: "POST", credentials: "include" });
    } catch {
    }
    clearScopes();
    message.success("Logged out successfully");
    setTimeout(() => navigate("/login"), 800);
  };

  const drawer = (
    <Box
      sx={{
        display: "flex",
        flexDirection: "column",
        height: "100%",
        overflow: "hidden",
      }}
    >
      <Toolbar sx={{ px: 3 }}>
        <Typography variant="h6" noWrap component="div" fontWeight={600}>
          ShortMesh
        </Typography>
      </Toolbar>
      <Divider sx={{ mx: 2, mb: 1 }} />
      <List sx={{ flex: 1, px: 1 }}>
        {menuItems.map((item) => {
          const accessible = hasResourceScope(item.resource);
          return (
            <ListItem key={item.text} disablePadding sx={{ mb: 0.5 }}>
              <ListItemButton
                selected={location.pathname === item.path}
                onClick={() => {
                  if (!accessible) {
                    message.info(`You do not have access to ${item.text}. Contact admin.`);
                    return;
                  }
                  navigate(item.path);
                }}
                sx={{
                  borderRadius: 2,
                  opacity: accessible ? 1 : 0.45,
                  "&.Mui-selected": {
                    backgroundColor: "background.default",
                    "&:hover": { backgroundColor: "primary.default" },
                  },
                }}
              >
                <ListItemIcon sx={{ color: accessible ? "inherit" : "text.disabled" }}>
                  {item.icon}
                </ListItemIcon>
                <ListItemText
                variant="body2"
                  primary={item.text}
                  primaryTypographyProps={{ variant: "body2" }}
                  sx={{ color: accessible ? "inherit" : "text.disabled" }}
                />
              </ListItemButton>
            </ListItem>
          );
        })}
      </List>
       <Divider sx={{ mx: 2, mt: 1 }} />
      <List sx={{ px: 1, pb: 2 }}>
        <ListItem disablePadding>
          <ListItemButton
            onClick={handleLogout}
            sx={{
              borderRadius: 2,
              "&:hover": {
                backgroundColor: "error.dark",
              },
            }}
          >
            <ListItemIcon>
              <Logout />
            </ListItemIcon>
            <ListItemText primary="Logout" />
          </ListItemButton>
        </ListItem>
      </List>
    </Box>
  );

  return (
    <Box
      sx={{
        display: "flex",
        height: "100vh",
        backgroundColor: "background.default",
      }}
    >
      <AppBar
        position="fixed"
        sx={{
          width: { sm: `calc(100% - ${drawerWidth + 16}px)` },
          ml: { sm: `${drawerWidth + 16}px` },
          display: { sm: "none" },
        }}
      >
        <Toolbar>
          <IconButton
            color="inherit"
            edge="start"
            onClick={handleDrawerToggle}
            sx={{ mr: 2, display: { sm: "none" } }}
          >
            <MenuIcon />
          </IconButton>
          <Typography variant="h6" noWrap component="div">
            ShortMesh Admin
          </Typography>
        </Toolbar>
      </AppBar>
      <Box
        component="nav"
        sx={{ width: { sm: drawerWidth + 16 }, flexShrink: { sm: 0 } }}
      >
        <Drawer
          variant="temporary"
          open={mobileOpen}
          onClose={handleDrawerToggle}
          ModalProps={{ keepMounted: true }}
          sx={{
            display: { xs: "block", sm: "none" },
            "& .MuiDrawer-paper": {
              boxSizing: "border-box",
              width: drawerWidth,
            },
          }}
        >
          {drawer}
        </Drawer>
        <Drawer
          variant="permanent"
          sx={{
            display: { xs: "none", sm: "block" },
            "& .MuiDrawer-paper": {
              boxSizing: "border-box",
              width: drawerWidth,
              margin: 2,
              marginRight: 0,
              height: "calc(100vh - 32px)",
              borderRadius: 3,
              border: "none",
              backgroundColor: "background.paper",
              boxShadow: "0 8px 32px rgba(0, 0, 0, 0.4)",
            },
          }}
          open
        >
          {drawer}
        </Drawer>
      </Box>
      <Box
        component="main"
        sx={{
          flexGrow: 1,
          p: 3,
          width: { sm: `calc(100% - ${drawerWidth + 16}px)` },
          mt: { xs: 8, sm: 0 },
          overflow: "auto",
        }}
      >
        <Outlet />
      </Box>
    </Box>
  );
}
