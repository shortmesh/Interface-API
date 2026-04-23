import { useState, useEffect } from "react";
import { Box, Typography, Button as MuiButton } from "@mui/material";
import {
  Add,
  Visibility,
  VisibilityOff,
  ContentCopy,
  Delete,
} from "@mui/icons-material";
import {
  Table,
  Modal,
  Input,
  Checkbox,
  Tag,
  Space,
  message,
  Alert as AntAlert,
  DatePicker,
  Button,
} from "antd";
import dayjs from "dayjs";
import {
  apiCall,
  safeJsonParse,
  maskString,
  formatDate,
  copyToClipboard,
} from "../utils/api";

export default function Tokens() {
  const [tokens, setTokens] = useState([]);
  const [loading, setLoading] = useState(true);
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [tokenDisplayDialogOpen, setTokenDisplayDialogOpen] = useState(false);
  const [useHost, setUseHost] = useState(false);
  const [setExpiry, setSetExpiry] = useState(false);
  const [expiryDate, setExpiryDate] = useState(null);
  const [attachToSession, setAttachToSession] = useState(false);
  const [displayToken, setDisplayToken] = useState("");
  const [revealedFields, setRevealedFields] = useState({});

  useEffect(() => {
    loadTokens();
  }, []);

  const loadTokens = async () => {
    try {
      const response = await apiCall("/api/v1/admin/tokens");
      if (!response) return;
      const data = await safeJsonParse(response);
      setTokens(Array.isArray(data) ? data : []);
    } catch (error) {
      console.error("Error loading tokens:", error);
      message.error("Failed to load tokens");
    } finally {
      setLoading(false);
    }
  };

  const handleCreateToken = async () => {
    try {
      const body = { use_host: useHost };
      if (setExpiry && expiryDate) {
        body.expires_at = expiryDate.toISOString();
      }

      const response = await apiCall("/admin/api/tokens", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      });

      if (!response) return;

      if (response.ok) {
        const data = await safeJsonParse(response);
        setDisplayToken(data.token);
        setCreateDialogOpen(false);
        setTokenDisplayDialogOpen(true);
        loadTokens();
      } else {
        const error = await safeJsonParse(response);
        message.error(error.error || "Failed to create token");
      }
    } catch (error) {
      console.error("Error creating token:", error);
      message.error("Failed to create token");
    }
  };

  const handleDeleteToken = async (id) => {
    Modal.confirm({
      title: "Delete Token",
      content: "Are you sure you want to delete this token?",
      okText: "Delete",
      okType: "danger",
      cancelText: "Cancel",
      onOk: async () => {
        try {
          const response = await apiCall(`/admin/api/tokens/${id}`, {
            method: "DELETE",
          });

          if (!response) return;

          if (response.ok) {
            loadTokens();
            message.success("Token deleted successfully");
          } else {
            const error = await safeJsonParse(response);
            message.error(error.error || "Failed to delete token");
          }
        } catch (error) {
          console.error("Error deleting token:", error);
          message.error("Failed to delete token");
        }
      },
    });
  };

  const handleCopy = async (text) => {
    try {
      await copyToClipboard(text);
      message.success("Copied to clipboard");
    } catch {
      message.error("Failed to copy");
    }
  };

  const toggleFieldVisibility = (fieldId) => {
    setRevealedFields((prev) => ({ ...prev, [fieldId]: !prev[fieldId] }));
  };

  const handleOpenCreateDialog = () => {
    setUseHost(false);
    setSetExpiry(false);
    setExpiryDate(null);
    setAttachToSession(false);
    setCreateDialogOpen(true);
  };

  const handleCloseTokenDisplay = async () => {
    if (attachToSession && displayToken) {
      try {
        await apiCall("/api/v1/admin/matrix-token", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ token: displayToken }),
        });
        message.success("Token attached to session");
      } catch (error) {
        console.error("Error attaching token:", error);
      }
    }
    setTokenDisplayDialogOpen(false);
    loadTokens();
  };

  useEffect(() => {
    if (setExpiry && !expiryDate) {
      const sevenDaysLater = dayjs().add(7, "day");
      setExpiryDate(sevenDaysLater);
    }
  }, [setExpiry]);

  const columns = [
    {
      title: "Username",
      dataIndex: "matrix_username",
      key: "username",
      render: (text, record) => (
        <Space size="small">
          <code style={{ fontFamily: '"Google Sans Mono", monospace' }}>
            {revealedFields[`username-${record.id}`] ? text : maskString(text)}
          </code>
          <Button
            type="text"
            size="small"
            icon={
              revealedFields[`username-${record.id}`] ? (
                <VisibilityOff style={{ fontSize: 16 }} />
              ) : (
                <Visibility style={{ fontSize: 16 }} />
              )
            }
            onClick={() => toggleFieldVisibility(`username-${record.id}`)}
          />
          <Button
            type="text"
            size="small"
            icon={<ContentCopy style={{ fontSize: 16 }} />}
            onClick={() => handleCopy(text)}
          />
        </Space>
      ),
    },
    {
      title: "Device ID",
      dataIndex: "matrix_device_id",
      key: "device_id",
      render: (text, record) => (
        <Space size="small">
          <code style={{ fontFamily: '"Google Sans Mono", monospace' }}>
            {revealedFields[`device-${record.id}`] ? text : maskString(text)}
          </code>
          <Button
            type="text"
            size="small"
            icon={
              revealedFields[`device-${record.id}`] ? (
                <VisibilityOff style={{ fontSize: 16 }} />
              ) : (
                <Visibility style={{ fontSize: 16 }} />
              )
            }
            onClick={() => toggleFieldVisibility(`device-${record.id}`)}
          />
          <Button
            type="text"
            size="small"
            icon={<ContentCopy style={{ fontSize: 16 }} />}
            onClick={() => handleCopy(text)}
          />
        </Space>
      ),
    },
    {
      title: "Admin",
      dataIndex: "is_admin",
      key: "is_admin",
      render: (isAdmin) => (
        <Tag color={isAdmin ? "warning" : "default"}>
          {isAdmin ? "Yes" : "No"}
        </Tag>
      ),
    },
    {
      title: "Expires",
      dataIndex: "expires_at",
      key: "expires_at",
      render: (date) => formatDate(date),
    },
    {
      title: "Last Used",
      dataIndex: "last_used_at",
      key: "last_used_at",
      render: (date) => formatDate(date),
    },
    {
      title: "Created",
      dataIndex: "created_at",
      key: "created_at",
      render: (date) => formatDate(date),
    },
    {
      title: "Actions",
      key: "actions",
      align: "center",
      render: (_, record) => (
        <Button
          type="text"
          danger
          icon={<Delete style={{ fontSize: 18 }} />}
          onClick={() => handleDeleteToken(record.id)}
        />
      ),
    },
  ];

  return (
    <Box>
      <Box
        sx={{
          mb: 4,
          display: "flex",
          justifyContent: "space-between",
          alignItems: "flex-start",
        }}
      >
        <Box>
          <Typography variant="h5" gutterBottom>
            Matrix Tokens
          </Typography>
          <Typography color="text.secondary" paragraph>
            Manage Matrix identity tokens for device linking and messaging.
          </Typography>
        </Box>
        <MuiButton
          variant="contained"
          startIcon={<Add />}
          onClick={handleOpenCreateDialog}
          sx={{ mb: 3 }}
        >
          Create New Token
        </MuiButton>
      </Box>

      <Table
        columns={columns}
        dataSource={tokens}
        loading={loading}
        rowKey="id"
        pagination={{ pageSize: 10 }}
        locale={{ emptyText: "No tokens found" }}
      />

      <Modal
        title="Create New Token"
        open={createDialogOpen}
        onOk={handleCreateToken}
        onCancel={() => setCreateDialogOpen(false)}
        okText="Create Token"
      >
        <div style={{ marginTop: 16 }}>
          <div style={{ marginBottom: 16 }}>
            <Checkbox
              checked={useHost}
              onChange={(e) => setUseHost(e.target.checked)}
            >
              <div>
                <div>Use admin credentials</div>
                <div
                  style={{
                    fontSize: 12,
                    color: "rgba(255, 255, 255, 0.6)",
                    marginTop: 4,
                  }}
                >
                  Reuse existing admin identity. Leave unchecked for first token
                  or to create new credentials.
                </div>
              </div>
            </Checkbox>
          </div>

          <div style={{ marginBottom: 16 }}>
            <Checkbox
              checked={setExpiry}
              onChange={(e) => setSetExpiry(e.target.checked)}
            >
              Set an expiry date
            </Checkbox>
          </div>

          {setExpiry && (
            <div style={{ marginBottom: 16 }}>
              <DatePicker
                showTime
                value={expiryDate}
                onChange={setExpiryDate}
                style={{ width: "100%" }}
                placeholder="Select expiry date"
              />
            </div>
          )}

          <div>
            <Checkbox
              checked={attachToSession}
              onChange={(e) => setAttachToSession(e.target.checked)}
            >
              <div>
                <div>Attach to current session for device management</div>
                <div
                  style={{
                    fontSize: 12,
                    color: "rgba(255, 255, 255, 0.6)",
                    marginTop: 4,
                  }}
                >
                  This token will be available for device operations during your
                  session and will be deleted and cleared on logout
                </div>
              </div>
            </Checkbox>
          </div>
        </div>
      </Modal>

      <Modal
        title="Token Created"
        open={tokenDisplayDialogOpen}
        onCancel={handleCloseTokenDisplay}
        footer={[
          <Button
            key="copy"
            type="text"
            icon={<ContentCopy />}
            onClick={() => handleCopy(displayToken)}
          >
            Copy
          </Button>,
          <Button
            key="close"
            type="text"
            onClick={handleCloseTokenDisplay}
            style={{ color: "#4357AD" }}
          >
            Close
          </Button>,
        ]}
      >
        <div style={{ marginTop: 16 }}>
          <AntAlert
            message="Save this token now. You won't be able to see it again!"
            type="warning"
            showIcon
            style={{ marginBottom: 16 }}
          />
          <Input.TextArea value={displayToken} rows={3} readOnly />
        </div>
      </Modal>
    </Box>
  );
}
