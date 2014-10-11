package com.ks.gopush.cli.model.request;

import java.io.Serializable;
import java.util.List;
import java.util.Map;



/**
 * 消息请求对象
 * 格式如下：
 * {
  "baseRequest" : {
       "token": "eflow_token"
       },
  "content": "Test!",
  "msgType": "1",
  "toUserNames" : ["22622391649384526@user", "22622391649384527@user"],
  "objectContent": {
      "appId": "23622391649370202",
      "appSendId": "xxxxx"
	  }
	}
 * @author ghg
 *
 */
public class MsgRequest implements Serializable {

	private static final long serialVersionUID = 1L;
   /**
    * 基础请求消息体
    */
	private BaseRequest baseRequest;
	/**
	 * 消息内容
	 */
	private String content = "Test!";
	/**
	 * 消息类型
	 */
	private String msgType = "1";
	/**
	 * 消息发送给那些用户
	 */
	private String[] toUserNames;
	/**
	 * 消息发送自定义内容
	 */
	private Map<String,String> objectContent;

	public BaseRequest getBaseRequest() {
		return baseRequest;
	}


	public void setBaseRequest(BaseRequest baseRequest) {
		this.baseRequest = baseRequest;
	}

	public String getContent() {
		return content;
	}

	public void setContent(String content) {
		this.content = content;
	}

	public String getMsgType() {
		return msgType;
	}

	public void setMsgType(String msgType) {
		this.msgType = msgType;
	}

	public String[] getToUserNames() {
		return toUserNames;
	}

	public void setToUserNames(String[] toUserNames) {
		this.toUserNames = toUserNames;
	}

	public Map<String, String> getObjectContent() {
		return objectContent;
	}

	public void setObjectContent(Map<String, String> objectContent) {
		this.objectContent = objectContent;
	}
}
