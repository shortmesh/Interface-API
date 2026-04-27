import { useState, useEffect, useRef } from "react";
import "../modalShake.css";
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
  App,
} from "antd";
import dayjs from "dayjs";
import {
  apiCall,
  safeJsonParse,
  maskString,
  formatDate,
  copyToClipboard,
} from "../utils/api";
import { CloseOutlined } from "@ant-design/icons";
import { hasScope } from "../utils/scopes";

export default function Tokens() {
  const { modal } = App.useApp();
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
  const [creating, setCreating] = useState(false);
  const [createTokenError, setCreateTokenError] = useState("");

  const [createShake, setCreateShake] = useState(false);
  const [displayShake, setDisplayShake] = useState(false);
  const [tokenVisible, setTokenVisible] = useState(false);

  const closeByXRef = useRef(false);

  const triggerShake = (setter) => {
    setter(true);
    setTimeout(() => setter(false), 500);
  };

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
    setCreating(true);
    setCreateTokenError("");
    try {
      const body = { use_host: useHost };
      if (setExpiry && expiryDate) {
        body.expires_at = expiryDate.toISOString();
      }

      const response = await apiCall("/api/v1/admin/tokens", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      });

      if (!response) return;

      if (response.ok) {
        const data = await safeJsonParse(response);
        setDisplayToken(data.token);
        setTokenVisible(false);
        setCreateDialogOpen(false);
        setTokenDisplayDialogOpen(true);
        loadTokens();
      } else {
        const error = await safeJsonParse(response);
        setCreateTokenError(error.error || "Failed to create token");
      }
    } catch (error) {
      console.error("Error creating token:", error);
      setCreateTokenError("Failed to create token");
    } finally {
      setCreating(false);
    }
  };

  const handleDeleteToken = async (id) => {
    modal.confirm({
      title: "Delete Token",
      content: "Are you sure you want to delete this token?",
      okText: "Delete",
      okType: "danger",
      cancelText: "Cancel",
      onOk: async () => {
        try {
          const response = await apiCall(`/api/v1/admin/tokens/${id}`, {
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
    setCreateTokenError("");
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
      render: (_, record) => {
        const canDelete = hasScope("tokens:write:delete");
        return (
          <Button
            type="text"
            danger
            icon={<Delete style={{ fontSize: 18 }} />}
            onClick={() =>
              canDelete
                ? handleDeleteToken(record.id)
                : message.info("You do not have permission to delete tokens. Contact admin.")
            }
            style={{ opacity: canDelete ? 1 : 0.45 }}
          />
        );
      },
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
          <Typography variant="h6" gutterBottom>
            Matrix Tokens
          </Typography>
          <Typography color="text.secondary" paragraph>
            Manage Matrix identity tokens for device linking and messaging.
          </Typography>
        </Box>
        {hasScope("tokens:write:create") ? (
          <MuiButton
            variant="contained"
            startIcon={<Add />}
            onClick={handleOpenCreateDialog}
            sx={{ mb: 3 }}
          >
            Create New Token
          </MuiButton>
        ) : (
          <MuiButton
            variant="contained"
            startIcon={<Add />}
            onClick={() => message.info("You do not have permission to create tokens. Contact admin.")}
            sx={{ mb: 3, opacity: 0.45 }}
          >
            Create New Token
          </MuiButton>
        )}
      </Box>

      {hasScope("tokens:read:*") && !hasScope("tokens:write:create") && (
        <AntAlert
          message="Read-only access"
          description="You can view tokens but don't have permission to create or delete them."
          type="info"
          showIcon
          style={{ marginBottom: 16 }}
        />
      )}

      <Table
        columns={columns}
        dataSource={tokens}
        loading={loading}
        rowKey="id"
        pagination={{ pageSize: 10 }}
        locale={{ emptyText: hasScope("tokens:read:*") ? "No tokens found" : "You do not have access to view tokens. Contact admin." }}
      />

      <Modal
        title="Create New Token"
        open={createDialogOpen}
        onCancel={() => { if (closeByXRef.current) { closeByXRef.current = false; setCreateDialogOpen(false); } else { triggerShake(setCreateShake); } }}
        maskClosable={true}
        closeIcon={<CloseOutlined onClick={() => { closeByXRef.current = true; }} />}
        wrapClassName={createShake ? 'modal-shake' : ''}
        footer={[
          <Button key="cancel" onClick={() => setCreateDialogOpen(false)} disabled={creating}>Cancel</Button>,
          <Button key="create" type="primary" onClick={handleCreateToken} loading={creating}>Create Token</Button>,
        ]}
      >
        <div style={{ marginTop: 16 }}>
          {createTokenError && (
            <AntAlert
              type="error"
              showIcon
              message={createTokenError}
              style={{ marginBottom: 16 }}
            />
          )}
          {hasScope("*") && (
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
          )}

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
        onCancel={() => { if (closeByXRef.current) { closeByXRef.current = false; handleCloseTokenDisplay(); } else { triggerShake(setDisplayShake); } }}
        maskClosable={true}
        closeIcon={<CloseOutlined onClick={() => { closeByXRef.current = true; }} />}
        wrapClassName={displayShake ? 'modal-shake' : ''}
        footer={[
          <Button
            key="close"
            type="text"
            onClick={handleCloseTokenDisplay}
            style={{ color: "#4357AD" }}
          >
            I have saved!
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
          <Box
            sx={{
              display: 'flex',
              alignItems: 'center',
              background: '#1e1e1e',
              border: '1px solid #333',
              borderRadius: 1,
              px: 1.5,
              py: 0.75,
              gap: 1,
            }}
          >
            <code
              style={{
                flex: 1,
                wordBreak: 'break-all',
                fontFamily: '"Google Sans Mono", monospace',
                fontSize: 13,
                color: '#e1e1e1',
                background: 'transparent',
              }}
            >
              {tokenVisible ? displayToken : maskString(displayToken)}
            </code>
            <Button
              type="text"
              size="small"
              style={{ color: '#aaa', flexShrink: 0 }}
              icon={tokenVisible ? <VisibilityOff style={{ fontSize: 16 }} /> : <Visibility style={{ fontSize: 16 }} />}
              onClick={() => setTokenVisible((v) => !v)}
            />
            <Button
              type="text"
              size="small"
              style={{ color: '#aaa', flexShrink: 0 }}
              icon={<ContentCopy style={{ fontSize: 16 }} />}
              onClick={() => handleCopy(displayToken)}
            />
          </Box>
        </div>
      </Modal>
    </Box>
  );
}
