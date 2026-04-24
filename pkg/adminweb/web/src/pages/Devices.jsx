import { useState, useEffect, useRef } from "react";
import { Link } from "react-router-dom";
import PhoneInput from "react-phone-number-input";
import "react-phone-number-input/style.css";
import "../phoneInput.css";
import "../modalShake.css";
import {
  Box,
  Typography,
  Button as MuiButton,
  Grid,
  Card,
  CardContent,
  CardActionArea,
  IconButton,
  Tooltip,
  Alert,
} from "@mui/material";
import {
  Add,
  Visibility,
  VisibilityOff,
  ContentCopy,
  Delete,
  Send,
} from "@mui/icons-material";
import { Table, Modal, Input, Tag, Space, message, Spin, Button } from "antd";
import { QRCodeSVG } from "qrcode.react";
import {
  apiCall,
  safeJsonParse,
  maskString,
  copyToClipboard,
} from "../utils/api";

export default function Devices() {
  const [devices, setDevices] = useState([]);
  const [loading, setLoading] = useState(true);
  const [addDeviceDialogOpen, setAddDeviceDialogOpen] = useState(false);
  const [sendMessageDialogOpen, setSendMessageDialogOpen] = useState(false);
  const [matrixTokenDialogOpen, setMatrixTokenDialogOpen] = useState(false);
  const [platformStep, setPlatformStep] = useState(true);
  const [qrCodeData, setQrCodeData] = useState(null);
  const [connectionStatus, setConnectionStatus] = useState("waiting");
  const [matrixToken, setMatrixToken] = useState("");
  const [matrixTokenError, setMatrixTokenError] = useState("");
  const [selectedDevice, setSelectedDevice] = useState(null);
  const [messageContact, setMessageContact] = useState("");
  const [messageText, setMessageText] = useState("");
  const [revealedFields, setRevealedFields] = useState({});
  const [platformLoading, setPlatformLoading] = useState(false);
  const [matrixTokenShake, setMatrixTokenShake] = useState(false);
  const wsRef = useRef(null);

  useEffect(() => {
    loadDevices();
    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, []);

  const loadDevices = async () => {
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

      const response = await apiCall("/api/v1/admin/devices");
      if (!response) return;
      const data = await safeJsonParse(response);
      setDevices(Array.isArray(data) ? data : []);
    } catch (error) {
      console.error("Error loading devices:", error);
      message.error("Failed to load devices");
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
        loadDevices();
      } else {
        const error = await safeJsonParse(response);
        setMatrixTokenError(error.error || "Failed to set token");
      }
    } catch (error) {
      console.error("Error setting token:", error);
      setMatrixTokenError("Failed to set token");
    }
  };

  const handleSelectPlatform = async (platform) => {
    setPlatformLoading(true);
    try {
      const response = await apiCall("/api/v1/admin/devices", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ platform }),
      });

      if (!response) return;

      if (!response.ok) {
        const error = await safeJsonParse(response);
        message.error(error.error || "Failed to add device");
        return;
      }

      const data = await safeJsonParse(response);
      const qrCodeUrl = data.qr_code_url;

      if (
        !qrCodeUrl ||
        (qrCodeUrl.includes("token=") && qrCodeUrl.endsWith("token="))
      ) {
        message.error("Failed to generate QR code");
        setAddDeviceDialogOpen(false);
        return;
      }

      setPlatformStep(false);

      setTimeout(() => {
        connectWebSocket(qrCodeUrl);
      }, 2000);
    } catch (error) {
      console.error("Error creating device:", error);
      message.error("Failed to add device");
    } finally {
      setPlatformLoading(false);
    }
  };

  const connectWebSocket = (qrCodeUrl) => {
    try {
      if (wsRef.current) {
        wsRef.current.close();
      }

      const ws = new WebSocket(qrCodeUrl);
      wsRef.current = ws;
      ws.receivedData = false;
      ws.hasError = false;

      ws.onopen = () => {
        setConnectionStatus("waiting");
      };

      ws.onmessage = (event) => {
        if (event.data.startsWith("Error:")) {
          ws.hasError = true;
          setConnectionStatus("error");
          message.error(event.data);
        } else {
          ws.receivedData = true;
          setQrCodeData(event.data);
        }
      };

      ws.onerror = () => {
        console.error("[Device Linking] WebSocket error");
        ws.hasError = true;
        setConnectionStatus("error");
        message.error("Connection error");
      };

      ws.onclose = () => {
        if (ws.receivedData && !ws.hasError) {
          setConnectionStatus("connected");
          message.success("Device added successfully");
          loadDevices();
          setTimeout(() => {
            setAddDeviceDialogOpen(false);
            resetAddDeviceDialog();
          }, 2000);
        }
      };
    } catch (error) {
      console.error("[Device Linking] Error connecting WebSocket:", error);
      setConnectionStatus("error");
    }
  };

  const resetAddDeviceDialog = () => {
    setPlatformStep(true);
    setQrCodeData(null);
    setConnectionStatus("waiting");
    setPlatformLoading(false);

    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
  };

  const handleDeleteDevice = (deviceId, platform) => {
    Modal.confirm({
      title: "Delete Device",
      content: "Are you sure you want to delete this device?",
      okText: "Delete",
      okType: "danger",
      cancelText: "Cancel",
      onOk: async () => {
        try {
          const response = await apiCall("/api/v1/admin/devices", {
            method: "DELETE",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ device_id: deviceId, platform }),
          });

          if (!response) return;

          if (response.ok) {
            loadDevices();
            message.success("Device deleted successfully");
          } else {
            const error = await safeJsonParse(response);
            message.error(error.error || "Failed to delete device");
          }
        } catch (error) {
          console.error("Error deleting device:", error);
          message.error("Failed to delete device");
        }
      },
    });
  };

  const handleOpenSendMessage = (device) => {
    setSelectedDevice(device);
    setMessageContact("");
    setMessageText("");
    setSendMessageDialogOpen(true);
  };

  const handleSendMessage = async () => {
    if (!messageContact.trim()) {
      message.error("Contact number is required");
      return;
    }
    if (!messageText.trim()) {
      message.error("Message is required");
      return;
    }

    try {
      const response = await apiCall(
        `/api/v1/admin/devices/${selectedDevice.device_id}/message`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            contact: messageContact,
            platform: selectedDevice.platform,
            text: messageText,
          }),
        },
      );

      if (!response) return;

      if (response.ok) {
        setSendMessageDialogOpen(false);
        message.success("Message queued successfully");
      } else {
        const error = await safeJsonParse(response);
        message.error(error.error || "Failed to send message");
      }
    } catch (error) {
      console.error("Error sending message:", error);
      message.error("Failed to send message");
    }
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

  const handleOpenAddDevice = async () => {
    const tokenStatusResponse = await apiCall(
      "/api/v1/admin/matrix-token-status",
    );
    if (!tokenStatusResponse) return;
    const tokenStatus = await safeJsonParse(tokenStatusResponse);

    if (!tokenStatus.has_matrix_token) {
      setMatrixTokenDialogOpen(true);
      return;
    }

    setAddDeviceDialogOpen(true);
  };

  const columns = [
    {
      title: "Platform",
      dataIndex: "platform",
      key: "platform",
      width: 120,
      render: (platform) => <Tag color="blue">{platform.toUpperCase()}</Tag>,
    },
    {
      title: "Device ID",
      dataIndex: "device_id",
      key: "device_id",
      render: (deviceId, record, index) => (
        <Space size="small">
          <Typography
            variant="body2"
            component="code"
            sx={{ fontFamily: '"Google Sans Mono", monospace' }}
          >
            {revealedFields[`device-${index}`]
              ? deviceId
              : maskString(deviceId)}
          </Typography>
          <Tooltip title="Toggle visibility">
            <IconButton
              size="small"
              onClick={() => toggleFieldVisibility(`device-${index}`)}
            >
              {revealedFields[`device-${index}`] ? (
                <VisibilityOff fontSize="small" />
              ) : (
                <Visibility fontSize="small" />
              )}
            </IconButton>
          </Tooltip>
          <Tooltip title="Copy">
            <IconButton size="small" onClick={() => handleCopy(deviceId)}>
              <ContentCopy fontSize="small" />
            </IconButton>
          </Tooltip>
        </Space>
      ),
    },
    {
      title: "Actions",
      key: "actions",
      width: 120,
      align: "center",
      render: (_, device) => (
        <Space size="small">
          <Tooltip title="Send message">
            <IconButton
              sx={{ color: "#e1e1e1" }}
              size="small"
              onClick={() => handleOpenSendMessage(device)}
            >
              <Send />
            </IconButton>
          </Tooltip>
          <Tooltip title="Delete">
            <IconButton
              color="error"
              size="small"
              onClick={() =>
                handleDeleteDevice(device.device_id, device.platform)
              }
            >
              <Delete />
            </IconButton>
          </Tooltip>
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
            Devices
          </Typography>
          <Typography color="text.secondary" paragraph>
            View and manage all devices connected to your Matrix identity.
          </Typography>
        </Box>
        <MuiButton
          variant="contained"
          startIcon={<Add />}
          onClick={handleOpenAddDevice}
        >
          Add Device
        </MuiButton>
      </Box>

      <Table
        columns={columns}
        dataSource={devices}
        loading={loading}
        rowKey={(record) => `${record.platform}-${record.device_id}`}
        pagination={false}
        locale={{ emptyText: "No devices found" }}
      />

      <Modal
        title="Add New Device"
        open={addDeviceDialogOpen}
        onCancel={() => {
          setAddDeviceDialogOpen(false);
          resetAddDeviceDialog();
        }}
        footer={[
          <Button
            key="cancel"
            type="text"
            onClick={() => {
              setAddDeviceDialogOpen(false);
              resetAddDeviceDialog();
            }}
          >
            Cancel
          </Button>,
        ]}
        width={600}
      >
        {platformStep ? (
          <Box sx={{ pt: 2 }}>
            <Typography color="text.secondary" gutterBottom>
              Select the platform for your new device
            </Typography>
            <Grid container spacing={2} sx={{ mt: 2 }}>
              {[
                {
                  platform: "wa",
                  name: "WhatsApp",
                  image: "/admin/whatsapp.png",
                },
                {
                  platform: "signal",
                  name: "Signal",
                  image: "/admin/signal.png",
                },
                {
                  platform: "telegram",
                  name: "Telegram",
                  image: "/admin/telegram.png",
                },
              ].map((item) => (
                <Grid item xs={12} sm={4} key={item.platform}>
                  <Card>
                    <CardActionArea
                      onClick={() => handleSelectPlatform(item.platform)}
                      disabled={platformLoading}
                    >
                      <CardContent sx={{ textAlign: "center", py: 4 }}>
                        <img
                          src={item.image}
                          alt={item.name}
                          style={{
                            width: 80,
                            height: 80,
                            objectFit: "contain",
                          }}
                        />
                        <Typography variant="h6" sx={{ mt: 2 }}>
                          {item.name}
                        </Typography>
                      </CardContent>
                    </CardActionArea>
                  </Card>
                </Grid>
              ))}
            </Grid>
            {platformLoading && (
              <Box sx={{ textAlign: "center", mt: 3 }}>
                <Spin size="default" />
                <Typography color="text.secondary" sx={{ mt: 1 }}>
                  Setting up device...
                </Typography>
              </Box>
            )}
          </Box>
        ) : (
          <Box sx={{ pt: 2, textAlign: "center" }}>
            <Typography variant="h6" gutterBottom>
              Scan QR Code
            </Typography>
            <Typography color="text.secondary" sx={{ mb: 3 }}>
              Use your device to scan this QR code
            </Typography>
            {qrCodeData ? (
              <Box sx={{ display: "flex", justifyContent: "center", mb: 2 }}>
                <QRCodeSVG value={qrCodeData} size={300} />
              </Box>
            ) : (
              <Box sx={{ py: 10 }}>
                <Spin size="large" />
              </Box>
            )}
            {connectionStatus === "waiting" && (
              <Space>
                <Spin size="small" />
                <Typography>Waiting for device...</Typography>
              </Space>
            )}
            {connectionStatus === "connected" && (
              <Alert severity="success">Device connected successfully!</Alert>
            )}
            {connectionStatus === "error" && (
              <Alert severity="error">
                Connection failed. Please try again.
              </Alert>
            )}
          </Box>
        )}
      </Modal>

      <Modal
        title="Set Matrix Token"
        open={matrixTokenDialogOpen}
        onCancel={() => {
          setMatrixTokenShake(true);
          setTimeout(() => setMatrixTokenShake(false), 500);
        }}
        closable={false}
        wrapClassName={matrixTokenShake ? "modal-shake" : ""}
        footer={[
          <Button
            key="skip"
            type="text"
            onClick={() => setMatrixTokenDialogOpen(false)}
          >
            Skip
          </Button>,
          <Button
            key="set"
            type="text"
            onClick={handleSetMatrixToken}
            style={{ color: "#4357AD" }}
          >
            Set Token
          </Button>,
        ]}
      >
        <Typography color="text.secondary" sx={{ mb: 2 }}>
          Please set your Matrix token to manage devices.
        </Typography>
        <Input
          size="large"
          placeholder="Matrix Token"
          value={matrixToken}
          onChange={(e) => {
            setMatrixToken(e.target.value);
            setMatrixTokenError("");
          }}
          status={matrixTokenError ? "error" : ""}
          autoFocus
        />
        {matrixTokenError && (
          <Typography
            color="error"
            variant="caption"
            sx={{ mt: 0.5, display: "block" }}
          >
            {matrixTokenError}
          </Typography>
        )}
        <Typography
          color="text.secondary"
          variant="caption"
          sx={{ mt: 2, display: "block" }}
        >
          Your token will be stored for this session and will be cleared when
          you log out. <br />
          Don't have a token?{" "}
          <Link
            to="/tokens"
            style={{ color: "#4357AD", textDecoration: "underline" }}
          >
            Create one here
          </Link>
        </Typography>
      </Modal>

      <Modal
        title="Send Message"
        open={sendMessageDialogOpen}
        onCancel={() => setSendMessageDialogOpen(false)}
        footer={[
          <Button
            key="cancel"
            type="text"
            onClick={() => setSendMessageDialogOpen(false)}
          >
            Cancel
          </Button>,
          <Button
            key="send"
            type="text"
            icon={<Send style={{ fontSize: 16 }} />}
            onClick={handleSendMessage}
            style={{ color: "#4357AD" }}
          >
            Send
          </Button>,
        ]}
      >
        <Box sx={{ mt: 2 }}>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
            Contact Number
          </Typography>
          <PhoneInput
            international
            defaultCountry="CM"
            value={messageContact}
            onChange={setMessageContact}
            placeholder="Enter phone number"
          />
        </Box>
        <Box sx={{ mt: 2 }}>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
            Message
          </Typography>
          <Input.TextArea
            rows={4}
            placeholder="Enter your message"
            value={messageText}
            onChange={(e) => setMessageText(e.target.value)}
          />
        </Box>
      </Modal>
    </Box>
  );
}
