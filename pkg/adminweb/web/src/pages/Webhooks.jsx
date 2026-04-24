import { useState, useEffect } from "react";
import { Link } from 'react-router-dom'
import { Box, Typography, Button as MuiButton } from "@mui/material";
import '../modalShake.css'
import { Add, Edit, Delete } from "@mui/icons-material";
import { Table, Modal, Input, Switch, Tag, Space, message, Spin, Button } from 'antd';
import { apiCall, safeJsonParse, formatDate } from "../utils/api";

export default function Webhooks() {
  const [webhooks, setWebhooks] = useState([]);
  const [loading, setLoading] = useState(true);
  const [addDialogOpen, setAddDialogOpen] = useState(false);
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [matrixTokenDialogOpen, setMatrixTokenDialogOpen] = useState(false);
  const [webhookUrl, setWebhookUrl] = useState("");
  const [editWebhook, setEditWebhook] = useState(null);
  const [editUrl, setEditUrl] = useState("");
  const [editActive, setEditActive] = useState(true);
  const [matrixToken, setMatrixToken] = useState("");
  const [matrixTokenError, setMatrixTokenError] = useState("");
  const [matrixTokenShake, setMatrixTokenShake] = useState(false);

  useEffect(() => {
    loadWebhooks();
  }, []);

  const loadWebhooks = async () => {
    try {
      const tokenStatusResponse = await apiCall(
        "/api/v1/admin/matrix-token-status",
      );
      if (!tokenStatusResponse) return;
      const tokenStatus = await safeJsonParse(tokenStatusResponse);

      if (!tokenStatus.has_matrix_token) {
        setMatrixTokenDialogOpen(true);
        setLoading(false);
        return;
      }

      const response = await apiCall("/api/v1/admin/webhooks");
      if (!response) return;

      if (!response.ok) {
        const error = await safeJsonParse(response);
        message.error(error.error || "Failed to load webhooks");
        setWebhooks([]);
        return;
      }

      const data = await safeJsonParse(response);
      setWebhooks(Array.isArray(data) ? data : []);
    } catch (error) {
      console.error("Error loading webhooks:", error);
      message.error("Failed to load webhooks");
    } finally {
      setLoading(false);
    }
  };

  const handleSetMatrixToken = async () => {
    if (!matrixToken.trim()) {
      setMatrixTokenError("Matrix token is required");
      return;
    }
    if (!matrixToken.startsWith("mt_")) {
      setMatrixTokenError("Token must start with mt_");
      return;
    }

    try {
      const response = await apiCall("/api/v1/admin/matrix-token", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ token: matrixToken }),
      });

      if (!response) return;

      if (response.ok) {
        setMatrixTokenDialogOpen(false);
        setMatrixToken("");
        setMatrixTokenError("");
        message.success("Matrix token set successfully");
        loadWebhooks();
      } else {
        const error = await safeJsonParse(response);
        setMatrixTokenError(error.error || "Failed to set token");
      }
    } catch (error) {
      console.error("Error setting token:", error);
      setMatrixTokenError("Failed to set token");
    }
  };

  const handleAddWebhook = async () => {
    if (!webhookUrl.trim()) {
      message.error("Please enter a webhook URL");
      return;
    }

    try {
      const response = await apiCall("/api/v1/admin/webhooks", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ url: webhookUrl }),
      });

      if (!response) return;

      if (response.ok) {
        setAddDialogOpen(false);
        setWebhookUrl("");
        loadWebhooks();
        message.success("Webhook added successfully");
      } else {
        const error = await safeJsonParse(response);
        message.error(error.error || "Failed to add webhook");
      }
    } catch (error) {
      console.error("Error adding webhook:", error);
      message.error("Failed to add webhook");
    }
  };

  const handleUpdateWebhook = async () => {
    if (!editUrl.trim()) {
      message.error("Please enter a webhook URL");
      return;
    }

    try {
      const response = await apiCall(
        `/api/v1/admin/webhooks/${editWebhook.id}`,
        {
          method: "PUT",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ url: editUrl, active: editActive }),
        },
      );

      if (!response) return;

      if (response.ok) {
        setEditDialogOpen(false);
        setEditWebhook(null);
        loadWebhooks();
        message.success("Webhook updated successfully");
      } else {
        const error = await safeJsonParse(response);
        message.error(error.error || "Failed to update webhook");
      }
    } catch (error) {
      console.error("Error updating webhook:", error);
      message.error("Failed to update webhook");
    }
  };

  const handleDeleteWebhook = async (id) => {
    Modal.confirm({
      title: 'Delete Webhook',
      content: 'Are you sure you want to delete this webhook?',
      okText: 'Delete',
      okType: 'danger',
      cancelText: 'Cancel',
      onOk: async () => {
        try {
          const response = await apiCall(`/api/v1/admin/webhooks/${id}`, {
            method: "DELETE",
          });

          if (!response) return;

          if (response.ok) {
            loadWebhooks();
            message.success("Webhook deleted successfully");
          } else {
            const error = await safeJsonParse(response);
            message.error(error.error || "Failed to delete webhook");
          }
        } catch (error) {
          console.error("Error deleting webhook:", error);
          message.error("Failed to delete webhook");
        }
      }
    });
  };

  const handleOpenEdit = (webhook) => {
    setEditWebhook(webhook);
    setEditUrl(webhook.url);
    setEditActive(webhook.active);
    setEditDialogOpen(true);
  };

  const handleOpenAdd = async () => {
    const tokenStatusResponse = await apiCall(
      "/api/v1/admin/matrix-token-status",
    );
    if (!tokenStatusResponse) return;
    const tokenStatus = await safeJsonParse(tokenStatusResponse);

    if (!tokenStatus.has_matrix_token) {
      setMatrixTokenDialogOpen(true);
      return;
    }

    setWebhookUrl("");
    setAddDialogOpen(true);
  };

  const columns = [
    {
      title: 'URL',
      dataIndex: 'url',
      key: 'url',
      width: '40%',
      ellipsis: true,
    },
    {
      title: 'Status',
      dataIndex: 'active',
      key: 'active',
      width: '10%',
      render: (active) => (
        <Tag color={active ? 'success' : 'default'}>
          {active ? 'Active' : 'Inactive'}
        </Tag>
      ),
    },
    {
      title: 'Created',
      dataIndex: 'created_at',
      key: 'created_at',
      width: '20%',
      render: (date) => formatDate(date),
    },
    {
      title: 'Updated',
      dataIndex: 'updated_at',
      key: 'updated_at',
      width: '20%',
      render: (date) => formatDate(date),
    },
    {
      title: 'Actions',
      key: 'actions',
      width: '10%',
      align: 'center',
      render: (_, record) => (
        <Space size="small">
          <Button
            type="text"
            icon={<Edit style={{ fontSize: 18 }} />}
            onClick={() => handleOpenEdit(record)}
            style={{ color: '#4357AD' }}
          />
          <Button
            type="text"
            icon={<Delete style={{ fontSize: 18 }} />}
            onClick={() => handleDeleteWebhook(record.id)}
            danger
          />
        </Space>
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
            Webhooks
          </Typography>
          <Typography color="text.secondary" paragraph>
            Manage webhook URLs to receive message notifications.
          </Typography>
        </Box>
        <MuiButton variant="contained" startIcon={<Add />} onClick={handleOpenAdd}>
          Add Webhook
        </MuiButton>
      </Box>

      <Table
        columns={columns}
        dataSource={webhooks}
        loading={loading}
        rowKey="id"
        pagination={{ pageSize: 10 }}
        locale={{ emptyText: 'No webhooks found' }}
      />

      <Modal
        title="Add Webhook"
        open={addDialogOpen}
        onOk={handleAddWebhook}
        onCancel={() => setAddDialogOpen(false)}
        okText="Add Webhook"
      >
        <div style={{ marginTop: 16 }}>
          <Input
            placeholder="https://example.com/webhook"
            value={webhookUrl}
            onChange={(e) => setWebhookUrl(e.target.value)}
            onPressEnter={handleAddWebhook}
          />
          <div style={{ marginTop: 8, fontSize: 12, color: 'rgba(255, 255, 255, 0.6)' }}>
            Enter a valid URL to receive webhook notifications
          </div>
        </div>
      </Modal>

      <Modal
        title="Edit Webhook"
        open={editDialogOpen}
        onOk={handleUpdateWebhook}
        onCancel={() => setEditDialogOpen(false)}
        okText="Save Changes"
      >
        <div style={{ marginTop: 16 }}>
          <Input
            placeholder="Webhook URL"
            value={editUrl}
            onChange={(e) => setEditUrl(e.target.value)}
            onPressEnter={handleUpdateWebhook}
          />
          <div style={{ marginTop: 16, display: 'flex', alignItems: 'center', gap: 8 }}>
            <span>Active:</span>
            <Switch
              checked={editActive}
              onChange={setEditActive}
            />
          </div>
        </div>
      </Modal>

      <Modal
        title="Set Matrix Token"
        open={matrixTokenDialogOpen}
        onCancel={() => {
          setMatrixTokenShake(true)
          setTimeout(() => setMatrixTokenShake(false), 500)
        }}
        closable={false}
        wrapClassName={matrixTokenShake ? 'modal-shake' : ''}
        footer={[
          <Button key="skip" type="text" onClick={() => setMatrixTokenDialogOpen(false)}>
            Skip
          </Button>,
          <Button key="set" type="text" onClick={handleSetMatrixToken} style={{ color: '#4357AD' }}>
            Set Token
          </Button>,
        ]}
      >
        <div style={{ marginTop: 16 }}>
          <div style={{ marginBottom: 8, color: 'rgba(255, 255, 255, 0.6)' }}>
            Please set your Matrix token to manage webhooks.
          </div>
          <Input
            placeholder="Matrix Token"
            value={matrixToken}
            onChange={(e) => {
              setMatrixToken(e.target.value);
              setMatrixTokenError("");
            }}
            status={matrixTokenError ? 'error' : ''}
            onPressEnter={handleSetMatrixToken}
          />
          {matrixTokenError && (
            <div style={{ marginTop: 4, color: '#ff4d4f', fontSize: 12 }}>
              {matrixTokenError}
            </div>
          )}
          <div style={{ marginTop: 8, fontSize: 12, color: 'rgba(255, 255, 255, 0.6)' }}>
            Your token will be stored for this session and will be cleared when you log out.
            <br />
            Don't have a token?{" "}
            <Link
              to="/tokens"
              style={{ color: "#4357AD", textDecoration: "underline" }}
            >
              Create one here
            </Link>
          </div>
        </div>
      </Modal>
    </Box>
  );
}
