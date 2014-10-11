package com.ks.gopush.cli.model.response;

import java.io.Serializable;

/**
 * 消息BaseResponse
 * 
 * @author ghg
 *
 */
public class BaseResponse implements Serializable{
	/**
	 * 
	 */
	private static final long serialVersionUID = 1L;

	private int ret;

	private String errMsg;

	public int getRet() {
		return ret;
	}

	public void setRet(int ret) {
		this.ret = ret;
	}

	public String getErrMsg() {
		return errMsg;
	}

	public void setErrMsg(String errMsg) {
		this.errMsg = errMsg;
	}

}
